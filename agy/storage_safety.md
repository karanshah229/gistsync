# Storage Safety: Concurrency & Atomicity

Gistsync ensures that your local state remains consistent and corruption-free, even when multiple processes are running simultaneously or if the system crashes during a write operation.

## 🔒 Concurrent-Safe State (File Locking)

To prevent "Lost Updates" (where two processes load the same file and overwrite each other's changes), Gistsync uses **Advisory Locking**.

-   **Mechanism**: Powered by `github.com/gofrs/flock`.
-   **Locking Pattern**: The `WithLock` wrapper implements a strict **Lock -> Load -> Mutate -> Save -> Unlock** sequence. 
-   **OS Integration**: Locks are managed by the Operating System kernel (macOS, Linux, and Windows).
-   **Crash Resilience**: If a process dies while holding a lock, the OS automatically releases the lock. No manual cleanup of "lock files" is ever required.

## ⚡ Atomic Writes (Persistence Integrity)

Gistsync prevents file corruption by ensuring that `state.json` is never in a "half-written" state.

-   **Pattern**: Write-to-temp-then-Rename.
    1.  The JSON data is serialized to a temporary file (`state.json.tmp`).
    2.  The OS performs an atomic `Rename` operation to replace the original `state.json`.
-   **Atomicity**: On modern file systems, the rename is an atomic operation. The file either points to the old data or the new data; there is no intermediate state.
-   **Failure Handling**: If the disk is full or a write fails, the original `state.json` remains untouched.

## 📁 Architecture 

These utilities are centralized in the `internal` package to separate persistence "plumbing" from core business logic:

-   `internal/storage.go`: Contains `WriteAtomic` and `WithFileLock`.
-   `core/state.go`: High-level state management that delegates to the storage utilities.

## ❓ Frequently Asked Questions (FAQ)

### What happens if the `gistsync` process dies while holding a lock?
The lock is **automatically unblocked**. Since we use OS-level advisory locking (via `flock`), the lock is tied to the process ID. If the process crashes or is killed, the Operating System kernel immediately releases the lock. No manual cleanup of `.lock` files is required.

### Can the `state.json` file still get corrupted?
It is extremely unlikely. By using an **Atomic Write** (writing to a `.tmp` file and then renaming), the original `state.json` is never partially overwritten. The OS ensures the rename is atomic—either the data is completely updated, or the original stays intact.

### Is this solution truly cross-platform?
Yes. We use `github.com/gofrs/flock`, which translates locking requests into the correct low-level system calls for macOS/Linux (`flock`) and Windows (`LockFileEx`).

### Why do we need `return nil` in the `WithLock` blocks?
In Go, closures require explicit return values. Since our `WithLock` pattern uses the closure to decide whether to commit or abort, we return `nil` for success. If an error is returned instead, the "Save" step is skipped to prevent writing invalid data.
