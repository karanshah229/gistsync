# gistsync

`gistsync` is a provider-agnostic file sync engine with GitHub Gists as the first provider implementation. It allows you to sync local files or folders to GitHub Gists with 2-way hash-based change detection.

## Prerequisites

1.  **Go**: Ensure Go 1.20+ is installed.
2.  **GitHub CLI (`gh`)**: The tool uses the `gh` CLI for authentication and API interaction.
    - Install: `brew install gh`
    - Login: `gh auth login`

## Build Instructions

To build the project from source:

```bash
go build -o gistsync main.go
```

This will produce a `gistsync` binary in the current directory.

## Running the App

### 1. Initialize Sync
To start tracking a file or folder:
```bash
./gistsync init path/to/file_or_folder
```

### 2. Manual Sync
To sync changes manually:
```bash
./gistsync sync path/to/file_or_folder
```

### 3. Check Status
To see what needs to be pushed or pulled:
```bash
./gistsync status
```

### 4. Continuous Sync (Watcher)
To automatically sync changes as you save (and check for remote updates every 60s):
```bash
./gistsync watch
```

## Internal Documentation
For more detailed information, see the `agy/` directory:
- [CLI API Guide](agy/cli_api.md)
- [Implementation Walkthrough](agy/walkthrough.md)
