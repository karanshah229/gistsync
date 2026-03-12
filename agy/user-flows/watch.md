# Watcher Flow (Background Sync)

The `watch` command starts a long-running background process that monitors the filesystem for changes and triggers synchronization automatically.

```mermaid
graph TD
    Start([gistsync watch]) --> Load[Load Config & State]
    Load --> InitEngine[Initialize Sync Engine]
    
    subgraph "Watcher Setup"
    InitEngine --> FSWatch[Initialize FSNotify Watcher]
    FSWatch --> AddPaths[Add Tracked Paths to Watchlist]
    end
    
    subgraph "Event Loop"
    AddPaths --> Wait{Waiting for Event...}
    Wait -- FS Event --> Debounce[Wait for Debounce Interval]
    Debounce --> Valid{Still Relevant?}
    Valid -- Yes --> TriggerSync[Execute Sync Engine]
    Valid -- No --> Wait
    
    TriggerSync --> Wait
    end
    
    subgraph "Notification"
    TriggerSync --> Log[Log Sync Activity]
    Log --> SystemNotify[Optional: OS Notification]
    end
```

### Key Parameters
- **WatchInterval**: Frequency of checking for changes (though mostly event-driven).
- **WatchDebounce**: Milliseconds to wait after the last change before triggering a sync, preventing redundant uploads during rapid edits.
- **State Resilience**: The watcher reloads state if it receives a SIGHUP or similar reload signal (implementation dependent).
