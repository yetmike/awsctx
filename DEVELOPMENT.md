# Development Guide

This document explains the internal structure and build process of `awsctx`.

## Project Structure

The project follows the standard Go project layout:

- `cmd/awsctx`: Main entry point for the application.
- `internal/awsctx`: Core logic, including configuration parsing, AWS config/credential manipulation, and UI handling.
- `shell/`: Shell completion scripts.
- `.github/workflows`: CI/CD pipelines (build and release).

### Core Logic (`internal/awsctx`)

- `config.go` & `ini.go`: Handles parsing and modifying AWS INI files (`~/.aws/config`, `~/.aws/credentials`).
- `profile.go`: Logic for listing and switching profiles.
- `region.go`: Logic for listing and switching regions.
- `cache.go`: Simple caching mechanism for state (previous profile/region).
- `fzf.go`: Integration with `fzf` for interactive selection.

## Building from Source

To build the binary:

```bash
make build
```

This will create `awsctx` in the root directory.

To install to `/usr/local/bin`:

```bash
sudo make install
```

## Running Tests

Unit tests cover configuration parsing and Profile/Region switching logic using mock files.

```bash
make test
```

## Release Process

We use **GoReleaser** and GitHub Actions to automate releases.

1.  **Tagging**: Create a semantic version tag (e.g., `v0.2.0`):
    ```bash
    git tag v0.2.0
    git push origin v0.2.0
    ```

2.  **Automation**: The GitHub Action `.github/workflows/release.yml` triggers on the tag.
3.  **GoReleaser**:
    - Builds binaries for multiple platforms (Linux, macOS, Windows).
    - Packages them into `.tar.gz` and `.zip` archives.
    - Generates `checksums.txt`.
    - Creates a GitHub Release with these artifacts.
    - (If configured) Pushes the updated formula to the Homebrew tap.

## Versioning

The version is injected at build time using `-ldflags`.
The `Makefile` and GoReleaser handle this automatically.

To manually build with a specific version:
```bash
go build -ldflags "-X github.com/yetmike/awsctx/internal/awsctx.Version=1.0.0" ./cmd/awsctx
```
