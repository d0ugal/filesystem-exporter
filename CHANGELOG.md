# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0 (2025-08-10)


### Bug Fixes

* add checkout step and issues permission to release-please workflow ([b54841b](https://github.com/d0ugal/filesystem-exporter/commit/b54841bc39e93b7619cd4d869f97133aa3cfa5d2))
* add missing error handling for df output parsing in volume collector ([c835d53](https://github.com/d0ugal/filesystem-exporter/commit/c835d53f5918e1e1bd18a5b3ecf03c371d6eaa8f))
* add missing version field to golangci-lint config ([07d507d](https://github.com/d0ugal/filesystem-exporter/commit/07d507d15a683b052ac4294ca40633b5b4de869a))
* resolve all golangci-lint issues ([6a4939b](https://github.com/d0ugal/filesystem-exporter/commit/6a4939bdb444489fda3bfe9ab88551b17f658b45))

## [Unreleased]

### Added
- Initial release of filesystem-exporter
- Filesystem metrics collection using `df` command
- Directory size metrics collection using `du` command
- Configurable collection intervals per filesystem/directory
- Prometheus metrics export
- Health check endpoint
- Structured logging with JSON and text formats
- Graceful shutdown handling
- Retry logic with exponential backoff
- Comprehensive configuration validation
- Docker support with multi-platform builds
- Example configurations for various use cases

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A


