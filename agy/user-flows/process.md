# Process Management Flow

The `process` command allows users to monitor and control background synchronization processes.

```mermaid
graph TD
    Start([gistsync process list/kill-others]) --> Scan[Scan System Process Table]
    
    subgraph "Process Identification"
    Scan --> Filter[Filter for 'gistsync' binaries]
    Filter --> Parse[Extract PID, PPID, StartTime, RSS, Cmdline]
    end
    
    subgraph "Command Logic"
    Parse --> Action{List or Kill?}
    
    Action -- List --> Table[Display Formatted Tabular Output]
    
    Action -- Kill --> FindOthers[Identify PIDs != Current PID]
    FindOthers --> Kill[Send OS Signal: SIGTERM/SIGKILL]
    Kill --> Report[Report Number of Processes Terminated]
    end
    
    Table --> Done([Done])
    Report --> Done
```

### Purpose
- **List**: Useful for checking if the `watch` command is running correctly in the background.
- **Kill-others**: Essential for resolving "zombie" processes or ensuring only one sync engine is active.
