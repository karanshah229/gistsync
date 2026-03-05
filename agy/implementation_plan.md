# Implementation Plan: gistsync

gistsync is a CLI application that syncs local files or folders to GitHub Gists, using a provider-agnostic engine with hash-based change detection.

## Proposed Changes

### Project Initialization
- Initialize Go module `github.com/karanshah229/gistsync` (or similar).
- Create directory structure: `cmd/`, `core/`, `providers/`, `watcher/`, `internal/`.
- Add basic `main.go` and `go.mod`.

### Core Logic (`core/`)
- `hashing.go`: SHA256 content hashing. [NEW]
- `models.go`: `File`, `Mapping` structs. [NEW]
- `state.go`: `State` management (saving/loading `state.json`). [NEW]
- `conflict.go`: Sync decision logic (3-way hash comparison). [NEW]
- `engine.go`: Orchestrates sync between local and remote. [NEW]

### Providers (`providers/`)
- `provider.go`: `Provider` interface definition. [NEW]
- `github_gist.go`: Implementation using `gh` CLI commands. [NEW]
- `gitlab.go`: Empty stub for future expansion. [NEW]

### CLI & UI (`cmd/`)
- `root.go`: Base command using `cobra`. [NEW]
- `init.go`: Initialize current directory for sync (state helper). [NEW]
- `sync.go`: Sync file or directory. [NEW]
- `status.go`: State status. [NEW]
- `remove.go`: Remove mapping. [NEW]
- `watch.go`: Foreground watcher using `fsnotify`. [NEW]

### Watcher (`watcher/`)
- `watcher.go`: Monitor local changes (fsnotify) and remote changes (polling). [MODIFY]
  - Implement periodic polling (e.g., every 60s).
  - Check GitHub API rate limits via `gh api rate_limit`.

### Internal Utilities (`internal/`)
- `configpath.go`: OS-appropriate config directory locator. [NEW]

## Verification Plan

### Automated Tests
- `go test ./core/...`: Test hashing, sync logic, and state management.
- Mock provider tests for the sync engine.

### Manual Verification
1. Run `gistsync init`.
2. Sync a file: `gistsync sync test.txt`. Verify gist created on GitHub.
3. Edit file locally, run `gistsync sync test.txt`. Verify gist updated.
4. Edit gist on GitHub (using browser), run `gistsync sync test.txt`. Verify local file pulled.
5. Create conflict (edit both), run `gistsync sync test.txt`. Verify conflict reported.
6. Test directory sync: `gistsync sync my_folder`.
7. Test `gistsync watch` for automatic background sync.
