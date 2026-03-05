# Gist Visibility Management

This document details the implementation of Public/Private Gist support in `gistsync`, focusing on the challenges of the GitHub API and the "Transactional" safety measures implemented.

## Overview

Users can now specify if a Gist should be public or private:
1.  **At Initialization**: Using `gistsync init --public` or `--private`.
2.  **Later**: Using `gistsync visibility <path> --public` or `--private`.

## Technical Challenges

### 1. GitHub API Limitations
GitHub's API does **not** allow changing the visibility of an existing Gist via a simple update (PATCH).
-   **Secret → Public**: Ignored by the API.
-   **Public → Secret**: Forbidden.

**Solution**: We implemented a **Recreation Flow**. To change visibility, the engine must:
1.  Fetch the content of the old Gist.
2.  Create a brand new Gist with the target visibility.
3.  Delete the old Gist.

### 2. Transactional "All-or-Nothing" Safety
Since the recreation flow involves multiple remote and local steps, we use a "Distributed Transaction" pattern to prevent state desync:

| Step | Action | failure Result |
| :--- | :--- | :--- |
| **1. Create New** | Create the Gist on GitHub with new visibility. | **Safe**: No change to local state or old Gist. |
| **2. Commit State** | Update `state.json` with the **New Gist ID**. | **Rollback**: If saving state fails, we delete the new Gist. |
| **3. Clean Remote** | Delete the **Old Gist** from GitHub. | **Non-Critical**: State is already updated. User has a "leaked" old Gist but no data loss. |

This ensures that the `state.json` remains the Source of Truth and never points to a deleted Gist.

## Bugs & Fixes Found During Implementation

### 1. The "Ghost Flag" Bug
Initially, we used `--secret` for private gists. However, `gh gist create` considers "secret" the default and does not have a `--secret` flag.
-   **Fix**: Omit the flag for private gists; only use `--public` for public ones.

### 2. The Deletion Flag Mismatch
We used `gh gist delete -y`, but some environments require the long-form `--yes`.
-   **Fix**: Standardized on `--yes` to ensure reliable cleanup.

### 3. State Overwrite Race Condition
CLI commands were reloading state and calling `Save()` after the engine had already updated the disk. This caused new mappings to be "wiped out" by stale CLI data.
-   **Fix**: Refactored CLI commands to trust engine self-persistence and removed redundant `Save()` calls.

### 4. Sync-Aware Visibility
Recreating a Gist blindly from remote content could overwrite newer local changes.
-   **Fix**: `SetVisibility` now performs a full 3-way hash check (Local vs Remote vs Last-Synced) and uses the **latest** content for the new Gist. It also raises a `ConflictError` if both have changed.

## FAQ

**Q: Does it work for folders?**
A: Yes. The engine recursively reads the local folder and fetches the remote Gist structure to ensure the entire folder is recreated correctly.

**Q: What if I specify both --public and --private?**
A: The CLI will throw an error and exit before doing any work.

**Q: Is "Private" the default?**
A: Yes. If no flags are provided during `init` or if a file is auto-synced, it defaults to Private for maximum security.
