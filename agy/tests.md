# Testing Architecture & Best Practices

This document outlines the testing philosophy, architecture, and suites for `gistsync`.

## 🧠 Philosophy: User-Flow Driven Testing

We prioritize testing the **branches of user flows** as defined in the `agy/user-flows/` Mermaid diagrams. Our goal is not just line coverage, but functional coverage of every path a user can take through the CLI.

## 🏗️ Test Architecture

The test suite is split into two layers:

### 1. Go Unit Tests (`*_test.go`)
Used for verifying pure logic, state transitions, and edge cases in isolation.
- **`core/`**: Tests for hashing, conflict detection (`DetermineAction`), and state serialization.
- **`internal/`**: Tests for platform-specific logic like `autostart` (mocked), path repair, and config validation.
- **`pkg/`**: Tests for i18n and UI rendering components.

### 2. Bash Integration Tests (`tests/*.sh`)
Used for end-to-end verification of CLI commands and remote provider integration.
- **`common.sh`**: Shared utilities for setting up the environment, mocking `gh` (GitHub CLI), and cleanup.
- **`run_tests.sh`**: The master script that executes all Go and Bash test suites sequentially.

## 🧪 Key Test Suites

### Sync Flow (`test_sync.sh`)
Verifies the core 2-way sync engine.
- **Path Flattening**: Ensures nested directories are flattened correctly using `---`.
- **Action Verification**: Confirms `PUSH`, `PULL`, `NOOP`, and `CONFLICT` states.
- **Virtual State Projection**: Verifies that `state.json` is injected into Gist backups during config sync.

### Init Flow (`test_init.sh`)
Verifies environment setup and restoration.
- **Interactive Prompts**: Tests various branches of the interactive setup (Overwrite? Restore? Backup?).
- **Auth Simulation**: Mocks `gh auth` failures to ensure helpful instructions are displayed.
- **Restoration**: Simulates "Fresh Machine" scenarios by restoring config/state from a remote Gist and patching `PENDING` IDs.

### State Recovery (`test_recover.sh`)
Verifies the Write-Ahead Logging (WAL) and recovery system.
- **Reconstruction**: Deletes `state.json` and reconstructs it from logs.
- **Interruption Detection**: Verifies that syncs missing a `CHECKPOINT` log are recovered with a warning.
- **High Water Mark (HWM)**: Ensures recovery correctly ignores entries already committed to disk.

## 🛠️ Specialized Testing Techniques

### Shared Stdin Reader
To prevent input loss when piping multiple answers into the CLI, we use a shared `bufio.Reader` via `ui.GetSharedReader()`. This ensures that individual components don't "steal" each other's buffered input.

### Mocking Providers
We use the `gh` CLI's ability to respond to `XDG_CONFIG_HOME` to point it to temporary test directories. This allows us to test real Gist interactions (or mock them via filesystem links) without polluting the developer's actual GitHub account.

## 📝 How to Add New Tests

1. **Identify the Flow**: Consult the relevant Mermaid diagram in `agy/user-flows/`.
2. **Choose the Layer**:
   - Use **Go tests** if the logic depends only on inputs/outputs or internal state.
   - Use **Bash tests** if the logic involves CLI output, multiple commands, or file system side effects.
3. **Setup/Cleanup**: Always use `setup_test_env` in bash scripts and `t.TempDir()` in Go tests to avoid side effects.
4. **Assert**: Use `grep` for CLI output and `assert_gist_exists` for remote verification.
