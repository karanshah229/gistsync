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

## 📖 Technical Documentation
For deep-dives into the architecture:
- [Version Management Guide](agy/version_management.md)
- [Visibility & Transactional Safety Guide](agy/visibility_guide.md)
- [Build & Distribution System](agy/build_system.md)
- [Implementation Walkthrough](agy/walkthrough.md)
- [CLI API Guide](agy/cli_api.md)
