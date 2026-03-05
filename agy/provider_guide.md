# Gistsync Provider & Diagnostics Guide

`gistsync` relies on external providers (currently GitHub Gists) to sync your files. To ensure everything is working correctly, we've implemented a robust diagnostic system.

## 🔍 The Provider Test Command

The quickest way to check if your setup is healthy is by running:

```bash
gistsync provider test
```

This command performs a series of checks:

1.  **Dependency Check**: Verifies that the required CLI tools (like `gh` for GitHub) are installed and available in your `$PATH`.
2.  **Authentication Check**: Ensures you are logged in and authorized to manage Gists.
3.  **API Health**: Performs a lightweight call to check connectivity and fetch your current API rate limits.

### Example Output
```text
🔍 Testing Provider: GitHub
✅ Success: GitHub CLI is authenticated and ready.
📊 Rate Limit: 4999 remaining (resets at 18:45:12)
```

## 🛡️ Installation Safety Nets

When you run the [install.sh](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/scripts/install.sh) script, `gistsync` automatically runs a provider check at the end. 

If it detects that the GitHub CLI (`gh`) is missing or not authenticated, it will render a clear warning:

> ⚠ **WARNING**: GitHub CLI is installed but NOT authenticated.
> Please run `gh auth login` to use gistsync.

## 🛠️ Internal Architecture

For developers, the `core.Provider` interface now requires a `Verify()` method. This ensures that any future provider implementation (e.g., GitLab, Bitbucket) must provide its own diagnostic logic, keeping the user experience consistent.

```go
type Provider interface {
    // ... other methods ...
    Verify() (bool, string, error)
}
```

---

If `provider test` fails, check the error message for specific instructions on how to resolve the issue (e.g., logging in or installing dependencies). 🏔️
