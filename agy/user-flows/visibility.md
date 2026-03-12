# Visibility Transition Flow

Changing a Gist's visibility (public <-> private) requires creating a new Gist because providers like GitHub do not allow visibility toggling on existing Gists.

```mermaid
graph TD
    Start([gistsync visibility path --public/--private]) --> LoadState[(Load state.json)]
    LoadState --> FindMapping{Is Tracked?}
    
    FindMapping -- No --> Abort([Abort: Untracked])
    FindMapping -- Yes --> CheckSame{Change Needed?}
    
    CheckSame -- No --> Exit([Already at Visibility])
    CheckSame -- Yes --> FetchRemote[Fetch Remote Content]
    
    subgraph "Consistency Check"
    FetchRemote --> ReadLocal[Read Local Content]
    ReadLocal --> Conflict{Conflict?}
    Conflict -- Yes --> Error([Abort: Resolve Conflict First])
    Conflict -- No --> PickLatest[Select Latest Content]
    end
    
    subgraph "Transactional Update"
    PickLatest --> CreateNew[Create NEW Gist with Target Visibility]
    CreateNew --> UpdateState[(Update state.json: New ID & Visibility)]
    
    UpdateState -- Success --> DeleteOld[Delete OLD Gist]
    UpdateState -- Failure --> Rollback[Delete NEW Gist]
    end
    
    DeleteOld --> Done([Visibility Updated])
    Rollback --> Error
```

### Risk Mitigation
- **Atomic State Update**: The new Gist ID is only committed to local state if the entire transaction succeeds.
- **Rollback**: If saving the local state fails, the newly created Gist is automatically deleted to prevent orphaned "shadow" Gists.
- **Cleanup**: The old Gist is only deleted after the local state has been successfully updated with the new Gist's information.
