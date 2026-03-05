# Gistsync Version Management

This document details the "Single Source of Truth" versioning strategy used in `gistsync`.

## 1. The Source of Truth (`VERSION`)

The core of the system is the **[VERSION](file:///Users/karan/projects/Personal_projects/gh-gist-syncer/VERSION)** file in the root directory.
- This file contains the plain SemVer string (e.g., `0.0.7`).
- All build processes (local, dev, and CI/CD) derive their version from this file.

## 2. Binary Injection (`ldflags`)

To ensure the binary knows its own version even after being moved or renamed, we use Go's `-ldflags` to inject values at compile time.

### Local/Dev Builds
The `Makefile` injects the version from the file plus the short Git hash:
```bash
go build -ldflags="-X 'github.com/karanshah229/gistsync/cmd.version=$(VERSION)-$(GIT_HASH)'"
```

### Releases
GoReleaser injects the clean version (e.g., `0.0.8`) during the official release process.

## 3. Embedded Fallback (`go:embed`)

If a developer runs `go run main.go` directly (skipping the Makefile), the app still needs to know its version.
- We use **`//go:embed VERSION`** in `cmd/version.go` to bake the version file directly into the binary at compile time.
- The `init()` function in `cmd/version.go` checks if `version` was set by `ldflags`. If not, it falls back to the embedded string.

## 4. Interactive Guard (Git Hook)

To prevent "lazy" commits where the version is forgotten, we use an interactive **pre-commit hook**.

### Installation
```bash
make install-hooks
```

### Behavior
1. **Detection**: Checks if `VERSION` is changed in the current commit.
2. **Interactive Prompt**: If not changed, it asks the developer if they want to bump the version.
3. **SemVer Bumping**: Allows selecting Patch, Minor, or Major.
4. **Automated Changelog**: Prompts for a message and prepends it to `CHANGELOG.md` with the current date.
5. **Auto-staging**: Automatically `git add`s the modified `VERSION` and `CHANGELOG.md` files.

## 5. CI/CD Integration

The GitHub Actions workflow (`cicd.yml`) is now fully synchronized with this system:
- It reads the version from the `VERSION` file.
- It checks if a corresponding Git tag (e.g., `v0.0.7`) already exists.
- If it's a **new version**, it creates the tag and triggers the release.

---

This holistic approach ensures that `gistsync` always has a reliable, traceable version across all environments.
