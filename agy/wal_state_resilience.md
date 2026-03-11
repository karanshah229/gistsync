# WAL, State Resilience & Sync Output

This article documents the transactional WAL (Write-Ahead Log) system, state persistence refactor, virtual state projection repairs, watcher improvements, and sync output differentiation.

---

## 1. Transactional WAL Logging

### Event Types

| Event | When | Purpose |
|-------|------|---------|
| `SYNC_START` | Before any remote call | Marks transaction begin |
| `SYNC_SUCCESS` | After remote succeeds | Records remote ID, hash, provider |
| `SYNC_ERROR` | On remote failure | Records error message |
| `CHECKPOINT` | After `state.json` saved to disk | Confirms local commit |
| `RECOVERY_COMPLETE` | After `recover` finishes | Advances the High Water Mark |

### Transaction Flow

```
SyncStart → Provider.Update/Create → SyncSuccess → State.Save() → Checkpoint
```

The `CHECKPOINT` is logged **inside** `State.Save()` only after the atomic write completes. This ensures logs are self-sufficient for state reconstruction.

### Related Files
- [logger/logger.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/internal/logger/logger.go) — structured JSON logging
- [core/state.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/core/state.go) — `Save()` logs checkpoint
- [core/engine.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/core/engine.go) — transaction boundaries

---

## 2. Recovery (`gistsync recover`)

### How It Works

1. **Load baseline**: Read existing `state.json` (if any) into a recovery map
2. **Parse all logs**: Read every `.log` file, extract JSON entries
3. **HWM Pruning**: Find the latest `CHECKPOINT` or `RECOVERY_COMPLETE` entry
   - If `state.json` already has mappings, skip entries before the HWM
   - If `state.json` is empty/missing, **ignore the HWM** to allow full reconstruction
4. **Grouped discovery**: Group remaining entries by PID, find `SYNC_SUCCESS` without a matching `CHECKPOINT`
5. **Apply changes**: Add recovered mappings to state
6. **Finalize**: Save state, log `RECOVERY_COMPLETE` and `CHECKPOINT` to advance HWM

### Key Design Decision: HWM Skip for Empty State

When `state.json` is deleted (disaster scenario), the recovery must replay ALL log entries, not just those after the last checkpoint. The HWM is only applied when `initialCount > 0`.

```go
if initialCount > 0 {
    // Find HWM from CHECKPOINT/RECOVERY_COMPLETE
}
```

### Related Files
- [cmd/recover.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/cmd/recover.go)

---

## 3. State Persistence: `WithLock` Refactor

### The Problem: State Shadowing

**Before**: `WithLock` was a package-level function. It loaded fresh state, ran the callback, and saved. But the caller's `Engine.State` was never updated, so subsequent operations wrote stale in-memory state back to disk, **overwriting** the correct state.

```
Sync file A → WithLock saves state with A
Sync file B → Engine.State still has OLD state → WithLock saves state with B but WITHOUT A
```

### The Fix: `State.WithLock` Method

`WithLock` is now a **method on `*State`**. After saving, it synchronizes the caller's state object:

```go
func (s *State) WithLock(fn func(freshState *State) error) error {
    return storage.WithFileLock(path, func() error {
        freshState, _ := LoadState()          // Load from disk
        fn(freshState)                         // Apply changes
        freshState.Save()                      // Persist
        s.Mappings = freshState.Mappings       // Sync caller
        s.Version = freshState.Version
        return nil
    })
}
```

### Where `WithLock` Is Used

| Caller | What it updates |
|--------|----------------|
| `SyncFile` (Noop/Push/Pull) | `LastSyncedHash` |
| `SyncDir` (Noop/Push/Pull) | `LastSyncedHash` |
| `AddMapping` | Adds/overwrites mapping |
| `SetVisibility` | Recreates gist, updates RemoteID + Public |
| `cmd/remove.go` | Removes mapping |

### Rule: No Bare `State.Save()` After `WithLock`

If a function already uses `WithLock`, do NOT call `State.Save()` separately — `WithLock` handles persistence internally.

---

## 4. Virtual State Projection (Config Backup)

### The Circular Dependency

Syncing `~/.config/gistsync` creates a loop:
1. Sync updates the remote Gist
2. Engine updates `state.json` with new hash
3. `state.json` changed → looks like a new local change → triggers another sync

### The Solution

1. **Stable Hashing**: `ReadLocalDir` and `ComputeDirHash` exclude `state.json`, `.lock`, `.tmp`, and `logs/` via `IsIgnoredConfigFile`
2. **Projection**: Before pushing the config directory, the engine loads the **freshest** `state.json` from disk (not the potentially stale in-memory copy), updates the config mapping's hash, and injects this into the upload payload
3. **Fresh State Loading**: Both `SyncDir` and `initialSync` call `LoadState()` immediately before projection

```go
freshState, err := LoadState()  // Always fresh from disk
for i, m := range freshState.Mappings {
    if m.LocalPath == absPath {
        freshState.Mappings[i].LastSyncedHash = currentLocalHash
        break
    }
}
stateJSON, _ := json.MarshalIndent(freshState, "", "  ")
uploadFiles = append(uploadFiles, File{Path: "state.json", Content: stateJSON, ...})
```

### Status Command Consistency

`engine.Status` mirrors the same logic: if `DetermineAction` returns `NOOP` for the config directory, it manually compares the local `state.json` with the remote one. If they differ, it returns `PUSH`.

### Related Files
- [core/engine.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/core/engine.go) — `SyncDir`, `initialSync`, `Status`
- [agy/config_auto_backup.md](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/agy/config_auto_backup.md) — original design doc

---

## 5. Watcher: Config Directory Allowlist

### The Problem

The watcher was using a **blocklist** (`IsIgnoredConfigFile`) to filter events. This had two issues:
1. `.tmp` files created during atomic writes leaked through and triggered unnecessary syncs
2. Ignoring `state.json` entirely prevented the config backup from being triggered after a file sync

### The Fix: Allowlist Approach

For events inside the config directory, only `state.json` and `config.json` changes trigger a config backup sync. Everything else is silently ignored:

```go
if filepath.Dir(event.Name) == configDir {
    if name == storage.StateFileName || name == storage.ConfigFileName {
        lastPath = configDir
        debounceTimer.Reset(...)
    }
    continue  // All other config dir events are ignored
}
```

### The Double-Sync Cascade

When editing a tracked file:
1. Edit file → sync → `state.json` updated → watcher triggers config sync (#1)
2. Config sync → `state.json` updated again (new config hash) → watcher triggers config sync (#2)
3. Second config sync → NOOP (remote matches) → chain stops

The second sync is harmless (no API call, just a NOOP check).

### Related Files
- [watcher/watcher.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/watcher/watcher.go)

---

## 6. Ignored Config Files

These files are excluded from hashing and watcher events via `storage.IsIgnoredConfigFile`:

| File | Why |
|------|-----|
| `state.json` | Excluded from hash to prevent loops; handled specially by projection |
| `state.json.lock` | File lock artifact |
| `state.json.tmp` | Atomic write temp file |
| `config.json.tmp` | Atomic write temp file |
| `gistsync.log` | Legacy log file |
| `logs/` | Log directory (skipped as subdirectory) |

### Related Files
- [internal/storage/constants.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/internal/storage/constants.go)

---

## 7. Sync Output Differentiation

### Return Type Change

`SyncFile` and `SyncDir` now return `(SyncAction, error)` instead of just `error`. This lets callers display action-specific messages.

### Output Format

| Action | Single Path | SyncAll |
|--------|-------------|---------|
| NOOP | `⏸️  /path: already up to date` | Same |
| PUSH | `✅ ⬆️  /path: pushed to remote` | Same |
| PULL | `✅ ⬇️  /path: pulled from remote` | Same |

### Callers Updated

| File | How it uses the action |
|------|----------------------|
| `cmd/sync.go` | Displays action-specific message |
| `internal/sync_helper.go` | Displays action-specific message per mapping |
| `cmd/init.go` | Discards (`_, err := engine.SyncDir(...)`) |
| `watcher/watcher.go` | Discards (`_, err = w.Engine.SyncDir(...)`) |

### i18n Keys

```json
"SyncNoop": "⏸️  {{.Path}}: already up to date",
"SyncPushed": "⬆️  {{.Path}}: pushed to remote",
"SyncPulled": "⬇️  {{.Path}}: pulled from remote"
```

---

## 8. UI Formatting: `ui.Print` vs `ui.Printf`

The `status` command was displaying all files on one line because `ui.Printf` doesn't append a newline. Fixed by switching to `ui.Print` which does.

**Rule**: Use `ui.Print` for all user-facing output lines. Use `ui.Printf` only for inline prompts (like config input) where you don't want a trailing newline.

The `en.json` templates should **not** contain manual `\n` characters — let the print function handle line termination.

---

## FAQs

### Q: Why does `recover` show the same warnings twice?
**A**: Before the HWM fix, `recover` didn't log a `RECOVERY_COMPLETE` event after finishing. Without this marker, subsequent runs couldn't tell that recovery had already been applied. Now, `recover` logs `RECOVERY_COMPLETE` + `CHECKPOINT` to advance the HWM, and subsequent runs see "State is already up to date."

### Q: Why not just use `e.State` directly for virtual projection?
**A**: The in-memory `Engine.State` can be stale — it was loaded at command startup. If another sync just updated `state.json` on disk (e.g., syncing `test.txt` updated the mapping), the in-memory copy won't reflect that. Loading fresh from disk with `LoadState()` ensures the projection includes all recent changes.

### Q: Why does syncing the config directory cause `state.json` to update twice?
**A**: First update: the config sync itself writes its `LastSyncedHash`. Second update: there isn't one if NOOP. But if it's a PUSH, the `WithLock` call saves the new hash, which writes `state.json` on disk. The watcher sees this write and triggers one more cycle, which resolves to NOOP because the remote now matches.

### Q: How does the tool prevent infinite sync loops with the config directory?
**A**: Three mechanisms:
1. `ReadLocalDir` excludes `state.json` from hash computation (stable content hash)
2. Virtual projection ensures the remote `state.json` matches the post-sync local state
3. The watcher allowlist only lets `state.json` and `config.json` changes through — `.tmp` and `.lock` files are silently dropped

### Q: Why is the `logs/` directory excluded from config hashing?
**A**: Log files change with every command invocation. Including them would make the config directory hash unstable, causing a PUSH on every sync even when nothing meaningful changed.

### Q: Why does `SyncDir` Pull still use `e.State.Save()` in some places?
**A**: It doesn't anymore. All three cases (Noop, Push, Pull) now consistently use `e.State.WithLock()` to prevent state shadowing. This was caught during the audit.

### Q: Why use an allowlist instead of a blocklist for watcher events?
**A**: The blocklist approach (`IsIgnoredConfigFile`) required maintaining a growing list of filenames. Any new temp file pattern would need to be added. The allowlist approach is simpler and more robust: only explicitly named files (`state.json`, `config.json`) trigger syncs. Everything else is automatically ignored.

### Q: Can `state.json` be fully reconstructed from logs alone?
**A**: Yes. When `state.json` is deleted, `gistsync recover` replays all `SYNC_SUCCESS` entries from every log file and reconstructs the mappings. The HWM is skipped when the initial state is empty, ensuring a full replay.
