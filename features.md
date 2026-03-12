# Features of gistsync

`gistsync` is designed to be a comprehensive tool for managing local-to-Gist synchronization. Below is a detailed look at its capabilities.

## 🚀 Core Workflow

### `init`
Initialize synchronization for a local path.
- **Selective Sync**: Choose whether to sync as a single file or an entire directory.
- **Provider Selection**: Pick your preferred backend (currently GitHub Gist specialized).
- **Public/Secret**: Control the visibility of your upstream Gists from the start.

### `sync`
The workhorse command for manual synchronization.
- **Bidirectional**: Intelligently decides whether to PUSH (local is newer) or PULL (remote is newer) based on content hashes.
- **Targeted**: Sync a specific path or run `gistsync sync` with no arguments to sync everything.
- **No-Op Awareness**: Skips network requests if the local and remote states already match.

### `status`
Provides a high-level overview of your synchronization state.
- Lists all tracked paths.
- Shows short-hand remote IDs and provides a clear summary of what needs syncing.

## 🛠️ Management & Polish

### `watch` (Daemon Mode)
The "set and forget" feature.
- **Real-time**: Leverages OS-level file system events to detect changes instantly.
- **Debounced**: Clusters multiple rapid edits together to minimize API calls and prevent rate-limiting issues.
- **Low Resource**: Designed to sit quietly in the background with minimal CPU/Memory footprint.

### `autostart`
Ensure your sync engine starts whenever you do.
- **Cross-Platform**: Manages `launchd` on macOS and `systemd` on Linux.
- **Easy Lifecycle**: Commands like `gistsync autostart setup` and `gistsync autostart uninstall` make daemon management trivial.

### `visibility`
On-the-fly privacy management.
- Quick-toggle a tracked mapping between `public` and `secret`.
- Automatically handles the Gist recreation/migration required by the GitHub API.

## 🚑 Disaster Recovery & Maintenance

### `recover`
If you lose your `state.json` or move to a new machine without your configuration:
- **WAL-Powered**: Replays the Write-Ahead Logs to reconstruct your tracking history.
- **Full Replay**: Can build a state from scratch if only the logs remain.

### `repair`
Perfect for those who use multiple machines.
- **Path Reconciliation**: If you sync your config across OSes (e.g., macOS and Linux), `repair` fixed absolute path mismatches automatically.

### `config sync`
The ultimate safety net.
- `gistsync` can sync its *own* configuration directory (`~/.config/gistsync`) to a Gist.
- Includes your mappings and settings, making "new machine setup" as simple as a single pull.

---

> [!TIP]
> Use the `-v` or `--verbose` flag with any command to see detailed logs of what's happening under the hood.
