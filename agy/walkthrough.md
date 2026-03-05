# Walkthrough: gistsync Implementation

gistsync is a cross-platform CLI tool that synchronizes local files and folders with GitHub Gists using a hash-based change detection engine.

## Features Accomplished
- [x] **Provider-Agnostic Architecture**: Sync engine logic is decoupled from the GitHub-specific implementation.
- [x] **Hash-Based Detection**: Uses SHA256 hashes of file contents to detect local/remote changes and conflicts.
- [x] **2-Way Sync**: Supports pushing local changes to Gists and pulling remote changes to local files.
- [x] **Directory Support**: Recursively syncs entire folders as multi-file Gists.
- [x] **State Management**: Maintains local state in an OS-appropriate config directory (`~/.config/gistsync/state.json`).
- [x] **GitHub Integration**: Wraps the `gh` CLI for secure authentication and Gist API interactions.
- [x] **File Watcher**: Automated sync with debouncing using `fsnotify`.

## Project Structure
```text
gistsync/
├── cmd/             # Cobra CLI commands (init, sync, status, remove, watch)
├── core/            # Sync engine, hashing, models, and state management
├── internal/        # OS-specific path utilities
├── providers/       # GitHub (gh CLI wrapper) and GitLab (stub) implementations
├── watcher/         # fsnotify-based file watcher
├── main.go          # Entry point
└── go.mod           # Dependencies (cobra, fsnotify)
```

## Binary & Verification
The project was successfully built into a static binary:
- **Build Command**: `go build -o gistsync main.go`
- **Tests**: Unit tests for hashing and sync decision logic pass.
  - `go test ./core/...` -> `ok github.com/karanshah229/gistsync/core`

## How to Use
1. **Initialize Sync**: `gistsync init <file or directory>`
2. **Check Status**: `gistsync status [path]`
3. **Manual Sync**: `gistsync sync <path>`
4. **Automated Sync**: `gistsync watch` (runs in foreground)
5. **Stop Tracking**: `gistsync remove <path>`

## Technical Highlights
- **Circular Dependency Resolution**: Refactored the architecture to move the `Provider` interface into the `core` package, allowing providers to depend on `core` without creating a cycle.
- **Debounced Watcher**: The watcher waits 500ms after a change before triggering a sync to avoid excessive API calls during rapid edits.
- **Provider Multi-file Support**: The GitHub provider efficiently handles multi-file Gists by looping through files for updates or fetching them collectively via `gh gist view --json`.
