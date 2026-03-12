# Autostart Flow

The `autostart` command manages the automated launching of `gistsync` upon user login, utilizing platform-specific mechanisms.

```mermaid
graph TD
    Start([gistsync autostart enable/disable/status]) --> CheckOS{Identify OS}
    
    subgraph "Logic Branching"
    CheckOS -- macOS --> AppleScript[Use AppleScript for Login Items]
    CheckOS -- Linux --> DesktopEntry[Manage ~/.config/autostart/*.desktop]
    CheckOS -- Windows --> Registry[Manage Registry Run Keys]
    end
    
    subgraph "Commands"
    AppleScript & DesktopEntry & Registry --> Command{Command?}
    
    Command -- Status --> CheckExists[Check for Entry Existence]
    CheckExists --> ReportStatus([Report Enabled/Disabled])
    
    Command -- Enable --> Install[Create Startup Entry]
    Install --> UpdateConfig[(Update config.yaml: autostart=true)]
    UpdateConfig --> SuccessE([Success: Enabled])
    
    Command -- Disable --> Remove[Delete Startup Entry]
    Remove --> UpdateConfigD[(Update config.yaml: autostart=false)]
    UpdateConfigD --> SuccessD([Success: Disabled])
    end
```

### OS Implementation Details
- **macOS**: Uses `osascript` to add/remove login items for the current user.
- **Linux**: Creates a `.desktop` file in the user's autostart directory.
- **Windows**: (Future/MVP support) Interacts with the `CurrentVersion\Run` registry key.
- **Config Sync**: The autostart preference is saved in `config.yaml`, ensuring it can be synced across machines via `config sync`.
