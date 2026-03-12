# Architecture & Technical Design

`gistsync` is built with a "Safety-First" philosophy. It ensures that your data remains consistent and your local state never corrupts, even in the event of crashes or concurrent access.

## 🔒 Reliability & Safety

### Atomic Writes
We never overwrite the `state.json` file directly. Instead, we use a **Write-to-temp-then-rename** pattern.
1. Data is written to `state.json.tmp`.
2. The OS performs an atomic `Rename` to replace the original.
3. This ensures that even if power is lost during a write, the file is either completely updated or completely untouched—never "half-written."

### Advisory File Locking
To prevent "Lost Updates" when multiple processes (like the `watch` daemon and a manual CLI command) try to update the state simultaneously:
- **Lock -> Load -> Mutate -> Save -> Unlock** sequence.
- Uses OS-level `flock` (advisory locking), which is automatically released by the kernel if the process crashes.

### Transactional WAL (Write-Ahead Log)
Every significant event is recorded in a structured JSON log before or after it happens.
- **`SYNC_START`**: Marks the beginning of a remote transaction.
- **`SYNC_SUCCESS`**: Confirms the remote state was updated.
- **`CHECKPOINT`**: Logged **inside** the atomic write function after the local state hits the disk.
- This WAL allows the `recover` command to rebuild lost state with high precision.

## ⚡ Performance Optimizations

### API Batching
Instead of making one API call per file, `gistsync` batches all changes into a single `PATCH` or `POST` request. This dramatically reduces sync time for large directories and avoids hitting API rate limits.

### Intelligent Watcher Allowlist
The background watcher doesn't just watch everything; it uses an **allowlist**. It explicitly ignores its own lock files, temp files, and logs, ensuring that the act of syncing doesn't trigger a recursive sync loop.

### Rate Limit Awareness
The tool monitors your API quota. If you are dangerously close to being throttled (e.g., < 10 remaining requests), background polling slows down or pauses until your quota resets.

## 🧩 Advanced Logic

### Virtual State Projection
When syncing the tool's own configuration directory, a circular dependency exists: syncing updates `state.json`, which then needs syncing.
- **Solution**: The engine "projects" the upcoming state change *into* the upload payload.
- The remote Gist is updated with what the *post-sync* local state will look like, satisfying the hash check in a single pass.

### Stable Hashing
Content hashes for directories are computed by recursively hashing files while strictly excluding metadata that changes every run (like logs and locks). This ensures that "Status" remains clean and "No-Op" remains accurate.

---

> [!NOTE]
> The persistence layer is centralized in `internal/storage`, keeping the core sync logic in `core/engine.go` clean and focused on business rules.
