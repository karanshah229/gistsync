# gistsync

`gistsync` is a provider-agnostic file sync engine with GitHub Gists as the first provider implementation. It allows you to sync local files or folders to GitHub Gists with 2-way hash-based change detection.

## Prerequisites

1.  **GitHub CLI (`gh`)**: `gistsync` uses the `gh` CLI for authentication and API interaction.
    -   Install: `brew install gh`
    -   Login: `gh auth login`

## Installation

To install `gistsync` globally on your system:

### One-liner
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

## Running the App

### 1. Initialize Sync
To start tracking a file or folder:
```bash
gistsync init path/to/file_or_folder
```
-   **Visibility**: Use `--public` for public gists (Defaults to Private/Secret).
    ```bash
    gistsync init my_folder --public
    ```

### 2. Manual Sync
To sync changes manually:
```bash
gistsync sync path/to/file_or_folder
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
gistsync version
```

## Development Mode

For developers, `gistsync` includes a "React-like" live-reload mode:
```bash
make dev
```
This will watch your code and automatically recompile/install the system-wide command whenever you save.

## Internal Documentation
For more detailed information, see the `agy/` directory:
- [Visibility & Transactional Safety Guide](agy/visibility_guide.md)
- [Build & Distribution System](agy/build_system.md)
- [Implementation Walkthrough](agy/walkthrough.md)
- [CLI API Guide](agy/cli_api.md)
