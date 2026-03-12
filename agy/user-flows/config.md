# Config Management Flow

The `config` command suite handles user preferences, cloud backups, and cross-platform path migration.

## Get / Set Flow
```mermaid
graph TD
    Start([gistsync config get/set]) --> LoadConfig[(Load config.yaml)]
    LoadConfig --> ValidKey{Valid Key?}
    ValidKey -- No --> Error([Error: Unknown Key])
    ValidKey -- Yes --> Operation{Get or Set?}
    
    Operation -- Get --> Print[Print Value]
    Operation -- Set --> ValidVal{Valid Value?}
    ValidVal -- No --> ErrorVal([Error: Invalid Input])
    ValidVal -- Yes --> Update[(Update and Save config.yaml)]
    Update --> Success([Success])
```

## Config Sync (Backup/Restore)
```mermaid
graph TD
    StartSync([gistsync config sync]) --> LoadState[(Load state.json)]
    LoadState --> GetConfigDir[Identify Config Directory]
    GetConfigDir --> Engine[Initialize Sync Engine]
    Engine --> SyncDir[Execute SyncDir Flow]
    
    subgraph "Virtual State Projection"
    SyncDir --> Snapshot[Capture Fresh State]
    Snapshot --> Inject[Inject state.json into Gist Payload]
    Inject --> Upload[Update Remote Gist]
    end
    
    Upload --> Success([Backup Complete])
```

## Config Repair
```mermaid
graph TD
    StartRepair([gistsync config repair]) --> LoadState[(Load state.json)]
    LoadState --> Iterate[Iterate Mappings]
    
    subgraph "Path Normalization"
    Iterate --> Missing{Path Exists?}
    Missing -- No --> Search[Search for Relative Path Match]
    Search -- Found --> Update[Update Mapping to New Path]
    Search -- Not Found --> MarkMissing[Mark as MISSING]
    Missing -- Yes --> MarkValid[Mark as VALID]
    end
    
    MarkValid --> Iterate
    Update --> Iterate
    MarkMissing --> Iterate
    
    Iterate -- Done --> Save[(Save Updated state.json)]
```
