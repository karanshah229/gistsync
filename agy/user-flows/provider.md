# Provider Health Flow

The `provider` command allows users to verify their authentication and monitor API usage limits.

```mermaid
graph TD
    Start([gistsync provider github/gitlab test]) --> Init[Initialize Provider Client]
    Init --> Verify[provider.Verify]
    
    subgraph "Authentication Check"
    Verify --> AuthOk{Auth Valid?}
    AuthOk -- No --> Error([Error: Check PAT/Token])
    AuthOk -- Yes --> Success([Connection Verified])
    end
    
    subgraph "Rate Limit Check"
    Success --> RateLimit[provider.CheckRateLimit]
    RateLimit --> Print[Display Remaining Quota & Reset Time]
    end
    
    Print --> Done([Done])
```

## Provider Info
The `provider info` command simply prints a predefined guide for obtaining Personal Access Tokens (PATs) for both GitHub and GitLab, pointing users to the correct settings URLs.
