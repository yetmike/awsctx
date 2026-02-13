# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.0.2] - 2026-02-13

### Fixed
- GitHub Actions workflow trigger for automatic releases (`v*.*.*` pattern).
- Manual release trigger (`workflow_dispatch`) for debugging CI/CD.

## [0.0.1] - 2026-02-13

### Added
- **Complete Refactor**: Moved core logic to `internal/awsctx` and entry point to `cmd/awsctx`.
- **Automated Releases**: Integrated **GoReleaser** for building multi-platform binaries (Linux, macOS, Windows).
- **Distributions**: Added configurations for **Homebrew** and **Scoop** (community repositories).
- **Install Script**: Created `install.sh` for easy installation via `curl | bash`.
- **Documentation**: Comprehensive `README.md`, `DEVELOPMENT.md`, and project license.
- **Project Structure**: Standard Go project layout with `Makefile` for local builds.
