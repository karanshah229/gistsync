# Gistsync Configuration Guide

This document explains how to customize `gistsync` behavior using the configuration system.

## 📁 Storage Location

Configuration is stored in a JSON file at the following OS-standard location:

- **macOS**: `~/.config/gistsync/config.json`
- **Linux**: `~/.config/gistsync/config.json`
- **Windows**: `%AppData%\gistsync\config.json`

## ⚙️ Available Settings

| Setting | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `watch_interval_seconds` | Integer | `60` | How often the `watch` command polls GitHub for remote changes. |
| `watch_debounce_ms` | Integer | `500` | Delay after a local file change before triggering a sync (prevents multiple syncs during rapid saves). |
| `log_level` | String | `"info"` | Verbosity of the output. Options: `info`, `debug`, `error`. |

## 🛠️ Managing via CLI

You can manage these settings directly from your terminal:

### List All Settings
```bash
gistsync config list
```

### Get a Specific Value
```bash
gistsync config get watch_interval_seconds
```

### Set a Value
```bash
gistsync config set watch_interval_seconds 120
gistsync config set watch_debounce_ms 1000
gistsync config set log_level debug
```

### Manual Backup
You can manually back up your entire configuration directory (including file mappings) to your default provider:
```bash
gistsync config sync
```
*Note: This utilizes Virtual State Projection to prevent infinite sync loops.*

---

Changes to configuration (like `watch_interval_seconds`) take effect the next time you restart the `gistsync watch` command. 🏔️
