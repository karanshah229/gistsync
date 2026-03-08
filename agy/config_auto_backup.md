# Config Auto-Backup & Virtual State Projection

The tool implements a self-backing-up mechanism that ensures your configuration and file mappings are safely stored in a Gist.

## 🔄 The Circular Dependency Problem

Syncing the tool's own configuration directory (`~/.config/gistsync`) creates a "chicken-and-egg" problem:
1.  The sync updates the remote Gist.
2.  The engine then updates the local `state.json` with the new remote ID or hash.
3.  Because `state.json` changed, it looks like a new local change, triggering another sync.

## ✨ Virtual State Projection

To break this loop, we implement **Virtual State Projection**:

1.  **Stable Hashing**: The `ReadLocalDir` function ignores `state.json` and `.lock` files when calculating the content hash of the configuration directory. This makes the "content" of the backup stable.
2.  **Projection**: When pushing the configuration directory, the engine manually injects a "projected" version of the current `state.json` into the upload payload.
3.  **Consistency**: We update the `LastSyncedHash` for the config mapping *before* projecting it into the JSON. This ensures that after the sync, the remote `state.json` matches the local one exactly.

## 📡 Watcher Integration

The `watch` command is aware of this mechanism:
- It monitors `state.json` for changes (e.g., after syncing a regular file).
- When a state change occurs, it triggers a sync of the configuration directory.
- Because of the stable hashing and projection logic, this sync results in a `NOOP` if the remote is already consistent, effectively breaking any infinite sync loops.

## 🛡️ Atomic Safety

Even with projection, the tool maintains its strict safety standards:
- **Locking**: `state.json.lock` is always respected.
- **Atomic Writes**: `internal.WriteAtomic` ensures that `state.json` is never corrupted during the backup process.

---

*Related files*: [core/engine.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/core/engine.go), [watcher/watcher.go](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/watcher/watcher.go) 🏔️
