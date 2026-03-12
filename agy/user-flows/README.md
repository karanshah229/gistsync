# User Flows

This directory contains detailed technical flowcharts for each command in `gistsync`. These diagrams are intended to help developers and users understand the internal logic, state transitions, and provider interactions.

## Legend
- **Rectangle**: Process / Operation
- **Diamond**: Decision point
- **Cylinder**: Storage (local state or configuration)
- **Cloud**: Remote Provider (GitHub/GitLab)

## Flows

### Core Commands
- [**Init**](init.md): Setup, configuration, and state initialization.
- [**Sync**](sync.md): 2-way sync logic, conflict detection, and atomic updates.
- [**Status**](status.md): Change detection and status reporting.

### Management Commands
- [**Config**](config.md): Configuration settings, backup/restore, and path repair.
- [**Provider**](provider.md): Testing and verifying provider connections.
- [**Visibility**](visibility.md): Changing gist visibility (public/private).
- [**Remove**](remove.md): Stopping tracking of local paths.

### System & Background
- [**Watch**](watch.md): Background file watching and auto-sync trigger.
- [**Autostart**](autostart.md): Login-time automation across OS platforms.
- [**Process**](process.md): Managing active `gistsync` instances.
- [**Recover**](recover.md): Reconstructing state from WAL logs.
