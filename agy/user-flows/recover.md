# State Recovery (WAL Replay) Flow

The `recover` command uses a Write-Ahead Logging (WAL) strategy to reconstruct a corrupted or missing `state.json` from historical activity logs.

```mermaid
graph TD
    Start([gistsync recover]) --> ScanLogs[Scan ~/.gistsync/logs/*.log]
    
    subgraph "Log Parsing"
    ScanLogs --> FilterParse[Parse JSON Log Entries]
    FilterParse --> ChronoSort[Sort All Entries by Timestamp]
    end
    
    subgraph "Baseline & HWM"
    ChronoSort --> LoadBaseline[Load current state.json if exists]
    LoadBaseline --> FindHWM[Locate High Water Mark - Latest CHECKPOINT]
    end
    
    subgraph "Replay Engine"
    FindHWM --> FilterPostHWM[Filter for Entries > HWM]
    FilterPostHWM --> GroupPID[Group Entries by Process ID]
    GroupPID --> Replay[Process PID Groups Chronologically]
    
    Replay --> SyncSuccess{Event == SYNC_SUCCESS?}
    SyncSuccess -- Yes --> Apply[Update Mapping in Memory]
    SyncSuccess -- No --> Next[Skip / Internal Log]
    
    Apply --> Checkpoint{Followed by CHECKPOINT?}
    Checkpoint -- No --> Warn[Warn: Potential Interruption] --> Next
    Checkpoint -- Yes --> Next
    end
    
    Replay -- Done --> Save[(Atomic Write: New state.json)]
    Save --> LogRecovery[Log RECOVERY_COMPLETE]
```

### Technical Details
- **WAL Concept**: Every successful sync logs a `SYNC_SUCCESS` event containing the path, hash, and gist ID. These logs act as the source of truth for reconstruction.
- **HWM (High Water Mark)**: A `CHECKPOINT` log entry indicates that `state.json` was safely persisted to disk. Recovery only needs to replay events that happened *after* the last known checkpoint.
- **Interruption Detection**: If a `SYNC_SUCCESS` is logged without a subsequent `CHECKPOINT`, it implies the process crashed *before* saving the state. The user is warned, but the mapping is still applied as the provider state is likely updated.
