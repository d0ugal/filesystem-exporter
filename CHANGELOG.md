# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.2.1...v1.3.0) (2025-08-13)


### Features

* add dynamic version information to web UI and CLI ([dc4c65b](https://github.com/d0ugal/filesystem-exporter/commit/dc4c65b69c04f4df9ebb6a16001a8262ebe252c8))

## [1.2.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.2.0...v1.2.1) (2025-08-13)


### Bug Fixes

* **docker:** update Alpine base image to 3.22.1 for better security and reproducibility ([0bcedff](https://github.com/d0ugal/filesystem-exporter/commit/0bcedff69460b7df225631370c408c0650b81ac9))

## [1.2.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.1.5...v1.2.0) (2025-08-11)


### Features

* add environment variable configuration support ([c47fee8](https://github.com/d0ugal/filesystem-exporter/commit/c47fee84693fd56d9819805f977e236d3c7f5daa))

## [1.1.5](https://github.com/d0ugal/filesystem-exporter/compare/v1.1.4...v1.1.5) (2025-08-11)


### Bug Fixes

* release improvements ([01dd33f](https://github.com/d0ugal/filesystem-exporter/commit/01dd33fadf33a6231cd581f4cc5ab4a267450054))

## [1.1.4](https://github.com/d0ugal/filesystem-exporter/compare/v1.1.3...v1.1.4) (2025-08-11)


### Bug Fixes

* fold release into one workflow ([a540835](https://github.com/d0ugal/filesystem-exporter/commit/a54083501c137d1e5507fb1525e64ba3b38b307c))

## [1.1.3](https://github.com/d0ugal/filesystem-exporter/compare/v1.1.2...v1.1.3) (2025-08-10)


### Bug Fixes

* **ci:** use environment variables to avoid JavaScript syntax errors ([3c950ad](https://github.com/d0ugal/filesystem-exporter/commit/3c950adc30888b2e7d84d1c1a0c6e0355fa5df83))

## [1.1.2](https://github.com/d0ugal/filesystem-exporter/compare/v1.1.1...v1.1.2) (2025-08-10)


### Bug Fixes

* **ci:** resolve JavaScript syntax errors in GitHub Actions workflows ([b65ff5f](https://github.com/d0ugal/filesystem-exporter/commit/b65ff5ff3c1eb052a1824d359de7cecabda022c4))
* resolve critical memory leak in directory size calculation ([5c899f8](https://github.com/d0ugal/filesystem-exporter/commit/5c899f853f465df776e82e903dfca5fa3894e6ff))

## [1.1.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.1.0...v1.1.1) (2025-08-10)


### Bug Fixes

* improve df command parsing error handling ([477c963](https://github.com/d0ugal/filesystem-exporter/commit/477c963f4c20bdcc3728da58722709f091a22936))

## [1.1.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.0.0...v1.1.0) (2025-08-10)


### Features

* **build:** improve formatting and linting workflow ([037340e](https://github.com/d0ugal/filesystem-exporter/commit/037340e37abe1de73fccca8108a9f48c65779b2e))


### Bug Fixes

* **ci:** optimize golangci-lint configuration and add formatting ([4c93bd6](https://github.com/d0ugal/filesystem-exporter/commit/4c93bd63d0a9c907555895e6668514f918ab5a8f))
* **ci:** replace golangci-lint docker action with official action ([22bdd29](https://github.com/d0ugal/filesystem-exporter/commit/22bdd29fdf9e8d6c890eae83456992fd78138363))
* **ci:** update golangci-lint action to v8 ([a21a5e1](https://github.com/d0ugal/filesystem-exporter/commit/a21a5e1bdc150cd05bdabc09d5634dd8316bab66))
* **collectors:** pass context through all collector functions ([524f4af](https://github.com/d0ugal/filesystem-exporter/commit/524f4afc3befbe6e8cc50e32e0c4377836e50005))
* **deps:** update module github.com/gin-gonic/gin to v1.10.1 ([5ab3027](https://github.com/d0ugal/filesystem-exporter/commit/5ab30274408c195dece40593e17ef9b63d70131c))
* **deps:** update module github.com/gin-gonic/gin to v1.10.1 ([8511f9c](https://github.com/d0ugal/filesystem-exporter/commit/8511f9cc3229b315ebdf3820bed815d808bd42f3))
* **deps:** update module github.com/prometheus/client_golang to v1.23.0 ([dff2648](https://github.com/d0ugal/filesystem-exporter/commit/dff264869ceae5efd60aa77d5e7e7c92eb1af44e))
* **deps:** update module github.com/prometheus/client_golang to v1.23.0 ([dc93d69](https://github.com/d0ugal/filesystem-exporter/commit/dc93d694d654d08db3ea6b74871bb4e555b41e7c))
* **tests:** handle os.Remove errors in test cleanup ([eb696f3](https://github.com/d0ugal/filesystem-exporter/commit/eb696f3d68c715905b1225d396456a1f24460ddc))
* update dependencies to fix security vulnerabilities ([2cbb12d](https://github.com/d0ugal/filesystem-exporter/commit/2cbb12d3d1be645994a020cc7f627fc5604f015f))
* update golangci-lint configuration and server improvements ([6aafe69](https://github.com/d0ugal/filesystem-exporter/commit/6aafe69c3e3172e19206f0318eeeb6b4f8656332))
* use a minimal config ([6f1a107](https://github.com/d0ugal/filesystem-exporter/commit/6f1a1071edb0343c228bcd337eb78f45b11e9615))

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
