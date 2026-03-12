# Contributing to gistsync

Thank you for your interest in contributing to `gistsync`! This project aims to be a robust, reliable, and user-friendly tool for syncing files with GitHub Gists.

## 🚀 Setting Up Your Development Environment

### 1. Prerequisites
- **Go 1.24+**: Required for native tool management.
- **GitHub CLI (`gh`)**: Required for API authentication.

### 2. Initial Setup
Clone the repository and run the developer setup:
```bash
git clone https://github.com/karanshah229/gistsync.git
cd gistsync
make dev
```

## 🛠️ Development Workflow

### Hot Reloading
We use `air` for a React-like development experience. Run the following command to start the live-reload engine:
```bash
make dev
```
This will:
1. Pre-warm your `sudo` session (securely masked).
2. Watch for changes in `.go` and configuration files.
3. Automatically recompile and install the system-wide `gistsync` command on every save.

### Versioning & Changelog
We follow a strict "Single Source of Truth" versioning system using the `VERSION` file.

**Interactive Bumping**:
The project includes a Git pre-commit hook that will prompt you to bump the version if it hasn't been changed. It will ask:
1. **Bump type**: Major, Minor, or Patch.
2. **Changelog message**: A short summary of what changed.

The hook automatically updates `VERSION` and `CHANGELOG.md` and stages them for your commit.

### Automated Testing
We maintain a comprehensive testing suite consisting of both Go unit tests and Bash-based integration tests.

**Running all tests:**
```bash
make test
```

**Running specific suites:**
- **Go Unit Tests**: `go test ./...` (Covers core logic, hashing, conflict detection).
- **Integration Tests**: `./tests/run_tests.sh` (Covers CLI flows, sync states, and recovery).

For a deep dive into the testing architecture and best practices, see the [Testing Architecture Guide](agy/tests.md).

### Dependency Management
We use the following libraries for specialized functionality:
- `github.com/shirou/gopsutil/v3`: For cross-platform process management.
- `github.com/fsnotify/fsnotify`: For file system events.
- `github.com/AlecAivazis/survey/v2`: For interactive CLI prompts.

## 📖 Technical Documentation
For deep-dives into the architecture:
- [Init, Standardization & Testing Guide](agy/init_standardization_guide.md)
- [Version Management Guide](agy/version_management.md)
- [Visibility & Transactional Safety Guide](agy/visibility_guide.md)
- [Build & Distribution System](agy/build_system.md)
- [Implementation Walkthrough](agy/walkthrough.md)
- [CLI API Guide](agy/cli_api.md)
