# gistsync

`gistsync` is a provider-agnostic file sync engine with GitHub Gists as the first provider implementation. It allows you to sync local files or folders to GitHub Gists with 2-way hash-based change detection.

## Prerequisites

1.  **GitHub CLI (`gh`)**: `gistsync` uses the `gh` CLI for authentication and API interaction.
    -   Install: `brew install gh`
    -   Login: `gh auth login`

## Installation

To install `gistsync` globally on your system:

### One-liner (Recommended)
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/karanshah229/gistsync/main/scripts/install.sh)"
```

### From Source (Makefile)
```bash
git clone https://github.com/karanshah229/gistsync.git
cd gistsync
make install
```

### Go Install
```bash
go install github.com/karanshah229/gistsync@latest
```

## Usage

### 1. Initialize Tool
Before using `gistsync`, you must initialize the configuration and state:
```bash
gistsync init
```

- **Backup & Restore**: `init` can automatically search for and restore your configurations from GitHub Gists.

### 2. Sync File or Folder
To start tracking and syncing a path:
```bash
gistsync sync path/to/file_or_folder
```
-   **Initial Sync**: If the path is not yet tracked, it will create a new Gist.
-   **Visibility**: Use `--public` for public gists (Defaults to Private/Secret).
    ```bash
    gistsync sync my_folder --public
    ```
-   **Manual Mapping**: You can manually connect a local path to an existing remote Gist ID. `gistsync` will verify the file contents and alert you with a safe `CONFLICT` if they differ, preventing accidental overwrites. You can also specify a provider for the mapping using the `--provider` flag.
    ```bash
    gistsync sync my_file.txt <existing_gist_id> --provider github
    ```

### 3. Change Visibility
You can change the visibility of a tracked path at any time. This uses our **Transactional Engine** to safely recreate the Gist without losing your local state.
```bash
gistsync visibility path/to/file_or_folder --public
```

### 4. Continuous Sync (Watcher)
To automatically sync changes as you save (and check for remote updates every 60s):
```bash
gistsync watch
```

### 5. Check Status & Version
```bash
gistsync status
gistsync --version
```

### 6. Provider Diagnostics
```bash
gistsync provider github test
gistsync provider info
```

### 7. Autostart at Login
Enable or disable automatic sync (via `gistsync watch`) when you log in to your computer. Supports macOS (LaunchAgents), Linux (systemd), and Windows (Startup shortcuts).
```bash
gistsync autostart status
gistsync autostart enable
gistsync autostart disable
```

### 8. Process Management
Manage multiple running instances of `gistsync`.
```bash
gistsync process list          # Show all running gistsync processes
gistsync process kill-others   # Terminate other instances except the current one
```

## Uninstallation

To completely remove `gistsync` from your system:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/karanshah229/gistsync/main/scripts/uninstall.sh)"
```

## Contributing
Interested in making `gistsync` better? Please see our **[Contributing Guide](CONTRIBUTING.md)** 
