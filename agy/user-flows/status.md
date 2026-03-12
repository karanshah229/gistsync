# Status Flow

The `status` command provides a quick overview of the health and synchronization state of your tracked files.

```mermaid
graph TD
    Start([gistsync status path]) --> LoadState[(Load state.json)]
    LoadState --> FindPaths{Single Path or All?}
    
    subgraph "Status Collection"
    FindPaths -- Single --> CheckTracked{Is Tracked?}
    CheckTracked -- No --> Untracked([UNTRACKED])
    CheckTracked -- Yes --> ComputeLocal[Compute Local Hash]
    
    FindPaths -- All --> Loop[Loop through Mappings]
    Loop --> ComputeLocal
    
    ComputeLocal --> FetchRemote[Fetch Remote Gist]
    FetchRemote --> RemoteDeleted{Gist Missing?}
    RemoteDeleted -- Yes --> Deleted([REMOTE_DELETED])
    RemoteDeleted -- No --> ComputeRemote[Compute Remote Hash]
    
    ComputeRemote --> Compare[DetermineAction]
    end

    subgraph "Results"
    Compare -- Noop --> OK([UP TO DATE])
    Compare -- Push --> Local[LOCAL CHANGES]
    Compare -- Pull --> Remote[REMOTE CHANGES]
    Compare -- Conflict --> Conflict([CONFLICT])
    end
```

### Interpretation
- **UP TO DATE**: Local, Remote, and LastSynced hashes all match.
- **LOCAL CHANGES**: Local file was modified; next sync will PUSH.
- **REMOTE CHANGES**: Gist was modified online; next sync will PULL.
- **CONFLICT**: Both local and remote have changed since the last sync.
