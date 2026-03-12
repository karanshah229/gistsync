# gistsync CLI API Documentation

`gistsync` is a CLI tool for synchronizing local files and directories with GitHub Gists.

## Global Flags
- `-h`, `--help`: Help for gistsync or any sub-command.

## Commands

### `init [path]`
Initializes sync for a local file or directory. 
- If a folder is provided, it tracks the entire directory recursively as a multi-file Gist.
- Creates the initial Gist and saves the mapping in `state.json`.

### `sync [path]`
Performs a manually triggered 2-way sync between a local path and its remote Gist.
- **Push**: Local changes are uploaded if remote has not changed since the last sync.
- **Pull**: Remote changes are downloaded if local has not changed since the last sync.
- **Conflict**: If both local and remote have changed, the sync stops and reports a conflict.

### `status [path]`
Shows the current synchronization status relative to the remote Gist.
- `NOOP`: Everything is in sync.
- `PUSH`: Local has newer changes.
- `PULL`: Remote has newer changes.
- `CONFLICT`: Both sides have changes.
- `UNTRACKED`: Path is not currently mapped to any Gist.

### `remove [path]`
Stops tracking a specific local path.
- Removes the mapping from the local `state.json` file.
- Does *not* delete the remote Gist (for safety).

### `watch`
Starts a foreground process that monitors all tracked files and directories.
- **Local Monitoring**: Automatically triggers a sync when a local file change is detected (with 500ms debounce).
- **Remote Polling**: Checks GitHub for remote changes every 60 seconds.
- **Rate Limit Safety**: Automatically skips polling cycles if API rate limits (remaining requests) are too low.

### `config <subcommand>`
Manages tool configuration and backups.
- `list`: Shows all current configuration keys and values.
- `get <key>`: Prints the value of a specific configuration key.
- `set <key> <value>`: Updates a configuration key.
- `sync`: Manually backs up the entire configuration directory to a Gist provider. This uses **Virtual State Projection** to ensure the remote backup remains consistent and loop-free.

## State Management
State is stored in a JSON format at:
- macOS/Linux: `~/.config/gistsync/state.json`
- Windows: `%AppData%\gistsync\state.json`
