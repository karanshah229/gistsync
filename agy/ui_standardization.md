# UI Standardization and i18n Layer

This document outlines the standardized UI system and internationalization (i18n) layer implemented in `gistsync`.

## Overview

The tool uses a centralized UI package to ensure consistent formatting, colors, and multi-language support across all commands. Direct calls to `fmt.Printf` or `log` for user-facing messages are discouraged in favor of `pkg/ui`.

## Architecture

### 1. Internationalization (`pkg/i18n`)
- **Library**: `github.com/nicksnyder/go-i18n/v2`
- **Locales**: JSON files stored in `pkg/i18n/locales/`.
- **Initialization**: Automatically loaded on first use or via `i18n.Init()`.
- **Usage**:
  ```go
  translation := i18n.T("Key", map[string]interface{}{"Param": "Value"})
  ```

### 2. UI Layer (`pkg/ui`)
- **Library**: `github.com/fatih/color` for rich terminal output.
- **Icons**: Standardized icons for different message types:
  - ✅ **Success**: `ui.Success("Key", data)`
  - ❌ **Error**: `ui.Error("Key", data)`
  - 💡 **Info**: `ui.Info("Key", data)`
  - ⚠️ **Warning**: `ui.Warning("Key", data)`
- **Formatting**:
  - `ui.Header("Key", data)`: Prints a bold header with a prefix.
  - `ui.Confirm("Key", data)`: Handles interactive y/N prompts.

## Implementation Guidelines

1. **Add Keys**: Always add new messages to `pkg/i18n/locales/en.json`.
2. **Use UI Package**: Replace all success/error messages with `ui.Success` or `ui.Error`.
3. **Template Data**: Pass dynamic values using a map:
   ```go
   ui.Success("SyncSuccess", map[string]interface{}{"Path": path})
   ```
4. **Icons**: Do not hardcode icons in your logic; let the UI package handle them for consistency.

## Example Sync Output
```bash
▶️ Syncing /path/to/file...
✅ /path/to/file synced successfully
```
