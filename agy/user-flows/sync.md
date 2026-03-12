# Sync Flow

The `sync` command is the core engine, performing 2-way hash-based synchronization between local files and remote Gists.

```mermaid
graph TD
    Start([gistsync sync path]) --> LoadState[(Load state.json)]
    LoadState --> FindMapping{Is Path Tracked?}

    subgraph "Initial Sync (First time)"
    FindMapping -- No --> CreateGist{Create New Gist}
    CreateGist --> Upload[Upload Content]
    Upload --> AddMapping[(Add Mapping to state.json)]
    AddMapping --> Done([Done])
    end

    subgraph "2-Way Sync (Recurring)"
    FindMapping -- Yes --> FetchRemote[Fetch Remote Content]
    FetchRemote --> ComputeHashes[Compute Local & Remote Hashes]
    ComputeHashes --> DetermineAction{Determine Action}

    DetermineAction -- ActionNoop --> UpToDate([No Changes])
    
    DetermineAction -- ActionPush --> LocalNewer[Local is Newer]
    LocalNewer --> Push[Update Remote Gist]
    Push --> UpdateState[(Update state.json Hash)]
    UpdateState --> Done

    DetermineAction -- ActionPull --> RemoteNewer[Remote is Newer]
    RemoteNewer --> Pull[Update Local File/Dir]
    Pull --> UpdateState
    
    DetermineAction -- ActionConflict --> Conflict([Conflict Detected])
    Conflict --> UserDecision[Manual Resolution Required]
    end

    subgraph "Virtual State Projection (Backups)"
    Push -- If Config Dir --> InjectState[Inject CURRENT state.json into Backup]
    InjectState --> Push
    end
```

### Action Determination Logic
| Local == Remote | Local == Last | Remote == Last | Action |
| :--- | :--- | :--- | :--- |
| Yes | - | - | **NOOP** (Already in sync) |
| No | Yes | No | **PULL** (Remote changed) |
| No | No | Yes | **PUSH** (Local changed) |
| No | No | No | **CONFLICT** (Both changed) |
