# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.9.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.8.0...v1.9.0) (2025-08-20)


### Features

* **api:** add pretty JSON formatting for metrics info endpoint ([4750345](https://github.com/d0ugal/filesystem-exporter/commit/4750345cd0ea1ec016968f2186734956341b70fe))
* **ui:** improve layout with grid endpoints and reorder sections ([dc9cecc](https://github.com/d0ugal/filesystem-exporter/commit/dc9cecc71f817931ee1ccad38127de2cf398c3a0))

## [1.8.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.7.0...v1.8.0) (2025-08-19)


### Features

* **server:** add dynamic metrics information with collapsible interface ([936114a](https://github.com/d0ugal/filesystem-exporter/commit/936114ae9a8d59f8b29c4811ebe92ea7dcb511fb))


### Bug Fixes

* **lint:** pre-allocate slices to resolve golangci-lint prealloc warnings ([8bba4cc](https://github.com/d0ugal/filesystem-exporter/commit/8bba4cc379c1cc37525e5cfb88bc5269497fc4df))

## [1.7.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.6.4...v1.7.0) (2025-08-18)


### Features

* add interval metrics for better PromQL monitoring ([2e9035c](https://github.com/d0ugal/filesystem-exporter/commit/2e9035c21abe6d853148a2badb648d9e4c7a1e86))

## [1.6.4](https://github.com/d0ugal/filesystem-exporter/compare/v1.6.3...v1.6.4) (2025-08-17)


### Bug Fixes

* collect directories up to specified level (inclusive) ([37abe72](https://github.com/d0ugal/filesystem-exporter/commit/37abe723a7e8772d1a3ba86db3199585a9260421))
* **lint:** handle error return from os.RemoveAll in test cleanup ([155fdaf](https://github.com/d0ugal/filesystem-exporter/commit/155fdaf86c1475069d303f48f594ff557e97b896))

## [1.6.3](https://github.com/d0ugal/filesystem-exporter/compare/v1.6.2...v1.6.3) (2025-08-17)


### Bug Fixes

* **test:** remove hardcoded paths from TestLoadFromEnvDirectoriesWithColons ([343821e](https://github.com/d0ugal/filesystem-exporter/commit/343821e7aee9bee1e1653f9da5b83e264bb92e87))

## [1.6.2](https://github.com/d0ugal/filesystem-exporter/compare/v1.6.1...v1.6.2) (2025-08-17)


### Bug Fixes

* add nolint comments for contextcheck ([9e933e6](https://github.com/d0ugal/filesystem-exporter/commit/9e933e62fed129c5fea8bbbd1103e3a8e5785a72))

## [1.6.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.6.0...v1.6.1) (2025-08-17)


### Bug Fixes

* use background context for timeout to prevent cancellation propagation ([e98803c](https://github.com/d0ugal/filesystem-exporter/commit/e98803c738f3c3252a87a18494544137e58aab27))

## [1.6.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.5.1...v1.6.0) (2025-08-17)


### Features

* add available metrics section to web UI ([fbedc2b](https://github.com/d0ugal/filesystem-exporter/commit/fbedc2b8f4c6d3bd7c512831558aa5dd669d9f20))

## [1.5.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.5.0...v1.5.1) (2025-08-16)


### Bug Fixes

* update golangci-lint config to match working mqtt-exporter pattern ([508ff61](https://github.com/d0ugal/filesystem-exporter/commit/508ff61b4af1bb02f62c994a7361097de376b073))

## [1.5.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.4.1...v1.5.0) (2025-08-16)


### Features

* add automerge rules for minor/patch updates and dev dependencies ([6e9f0a1](https://github.com/d0ugal/filesystem-exporter/commit/6e9f0a1f5f5a4d9772def1bfdbb098a3555f9b4c))
* upgrade to Go 1.25 ([3dc1bb6](https://github.com/d0ugal/filesystem-exporter/commit/3dc1bb6e1ed68ef8b32e95d1e6456bcb5010d742))


### Bug Fixes

* revert golangci-lint config to version 2 for compatibility ([32285a6](https://github.com/d0ugal/filesystem-exporter/commit/32285a6b271f71259b7000cc855ad262b80abc73))
* update golangci-lint config for Go 1.25 compatibility ([15d2aed](https://github.com/d0ugal/filesystem-exporter/commit/15d2aeddf96347cacfac735ab1e33759643cee5b))

## [1.4.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.4.0...v1.4.1) (2025-08-14)


### Bug Fixes

* ensure correct version reporting in release builds ([5b40eb7](https://github.com/d0ugal/filesystem-exporter/commit/5b40eb70ec09a6ae832c9dd7f41695f6a6570efb))

## [1.4.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.3.0...v1.4.0) (2025-08-14)


### Features

* add version info metric and subtle version display in h1 header ([1df4595](https://github.com/d0ugal/filesystem-exporter/commit/1df4595ee9c77a5662ae40704086cb42310c2362))
* add version to title, separate version info, and add copyright footer with GitHub links ([b345a64](https://github.com/d0ugal/filesystem-exporter/commit/b345a643b106e671927d183c0e6a56d19787c452))


### Bug Fixes

* update Dockerfile to inject version information during build ([c28308e](https://github.com/d0ugal/filesystem-exporter/commit/c28308e91d81b4f9e735a357fcdabb5fb18927d3))

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
