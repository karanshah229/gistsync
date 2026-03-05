# Gistsync Build & Distribution System

This document outlines the build, versioning, and CI/CD infrastructure for `gistsync`.

## 1. Local Development (`Makefile`)

The `Makefile` provides a standardized entry point for building and installing `gistsync` on various operating systems.

-   **`make build`**: Compiles the binary locally, injecting Git version/hash into the binary using `-ldflags`.
-   **`make install`**: Automatically detects the OS and installs the binary to `/usr/local/bin` (macOS/Linux) or `go install` (Windows).
-   **`make clean`**: Removes locally built binaries and the `dist/` directory.

## 2. Versioning Strategy

Gistsync uses Semantic Versioning (SemVer) with automatic Git tagging.

-   **Base Version (`cmd.version`)**: The base version is stored in `cmd/version.go`. During build, it's injected via:
    ```bash
    go build -ldflags="-X 'github.com/karanshah229/gistsync/cmd.version=$(VERSION)'"
    ```
-   **Automatic Tagging**: Merges to the `main` branch trigger a GitHub Action (`anothrNick/github-tag-action`) that automatically increments the patch version (e.g., `v1.0.1` -> `v1.0.2`).

## 3. CI/CD Pipeline (`.github/workflows/cicd.yml`)

The GitHub Actions workflow is triggered on every push or PR to `main`.

1.  **Test**: Runs the unit test suite (`go test ./core/...`).
2.  **Tagging**: If the push is to `main`, a new version tag is calculated and pushed.
3.  **Release**: Triggers **GoReleaser** once the new tag is detected.

## 4. Multi-Platform Distribution (`.goreleaser.yaml`)

GoReleaser automates the creation of binaries for multiple platforms and architectures:

-   **Operating Systems**: Linux, macOS (Darwin), and Windows.
-   **Architectures**: Intel/AMD (`amd64`) and Apple Silicon/ARM (`arm64`).
-   **Archives**: Generates `.tar.gz` for Linux/macOS and `.zip` for Windows.
-   **Artifacts**: Includes `checksums.txt` for security and an auto-generated changelog.

---

This system ensures that building and releasing `gistsync` is reproducible, automated, and secure.
