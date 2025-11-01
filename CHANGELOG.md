# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.23.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.22.4...v1.23.0) (2025-11-01)


### Features

* add dev-tag Makefile target ([dcbc72a](https://github.com/d0ugal/filesystem-exporter/commit/dcbc72a472dbae9ed0978bb29cef0951e95315f7))
* add duplication linter (dupl) to golangci configuration ([0d8bee0](https://github.com/d0ugal/filesystem-exporter/commit/0d8bee019cff40ad51222c4055204a0c4529df2c))
* add tracing configuration support ([8da4627](https://github.com/d0ugal/filesystem-exporter/commit/8da462732d36cca14670f3872edafa8056136705))
* **ci:** add auto-format workflow ([0918747](https://github.com/d0ugal/filesystem-exporter/commit/09187470fe17835798af58436efc2d2e3a2a2694))
* integrate OpenTelemetry tracing into collectors ([8a88edc](https://github.com/d0ugal/filesystem-exporter/commit/8a88edc7c378fe30f85f8e55609720cff8aedd2c))
* **tracing:** add comprehensive tracing throughout filesystem-exporter ([ee0e576](https://github.com/d0ugal/filesystem-exporter/commit/ee0e576d693432945561add5691de5491f5aa894))
* trigger CI after auto-format workflow completes ([14bae61](https://github.com/d0ugal/filesystem-exporter/commit/14bae614a542f0ac85bb8a97fa714e849093424f))


### Bug Fixes

* add nolint comment for contextcheck and fix import ordering ([eae68bc](https://github.com/d0ugal/filesystem-exporter/commit/eae68bc02884937ecba1993c913ab87ee9743a27))
* add nolint comment for contextcheck on span context extraction ([e593073](https://github.com/d0ugal/filesystem-exporter/commit/e593073bca1cdbe3553a4ff35997ce94a3bfe433))
* add nolint comment for contextcheck on spanCtx variable declaration ([3577c73](https://github.com/d0ugal/filesystem-exporter/commit/3577c7323c8cbc7ef7e060159739cbdfc23d769a))
* add nolint comment for contextcheck on WithTimeout call ([2a91838](https://github.com/d0ugal/filesystem-exporter/commit/2a9183861bedf84fdce64d51b4bc8887bad25fde))
* add nolint comment to span.Context() assignments ([23415bb](https://github.com/d0ugal/filesystem-exporter/commit/23415bbbbd53a21127caf41bebbeb028b2425488))
* **config:** apply tracing env vars when loading from config file ([216fde0](https://github.com/d0ugal/filesystem-exporter/commit/216fde0546a86c5d85084d7d87bc609c1e1914fb))
* ensure spanCtx is set to ctx when tracer is disabled ([cb85044](https://github.com/d0ugal/filesystem-exporter/commit/cb8504446bd80cf84044af63b3eb726c0002110c))
* **lint:** remove empty branch to fix staticcheck SA9003 warning ([6133b63](https://github.com/d0ugal/filesystem-exporter/commit/6133b63ddd13869b1a23426066b9e255ec352e0a))
* **lint:** resolve contextcheck and typecheck by context-only tracing\n\n- Refactor volume collector to use context-only spans (no parent span arg)\n- Add minimal nolint contextcheck where required in directory collector\n- Remove duplicate retryCtx and unused vars\n- Add helper in directory collector to restore previous behavior\n\nAll linters pass. ([0aa8ea8](https://github.com/d0ugal/filesystem-exporter/commit/0aa8ea8a30105544a7154499f49bbb8f82382dd5))
* **metrics:** correct CollectionFailedCounter label names ([bbe0db5](https://github.com/d0ugal/filesystem-exporter/commit/bbe0db50b1b7b8df9bf6a34b1c5b3b1a3f098e0c))
* move nolint comment above WithTimeout call ([1fa9470](https://github.com/d0ugal/filesystem-exporter/commit/1fa9470baf3689d70483e647e6de658e76b606f3))
* pass BaseConfig to app for tracing initialization ([6faa161](https://github.com/d0ugal/filesystem-exporter/commit/6faa161d205714da2dbcc0839692044acdedfa51))
* remove unused nolint comments from assignments ([db8a020](https://github.com/d0ugal/filesystem-exporter/commit/db8a02071c65c7a8ab2bd0ce40d75b8637d68804))
* skip retries for non-retryable errors (signal: killed, context canceled) ([157d1e2](https://github.com/d0ugal/filesystem-exporter/commit/157d1e2bd717a1677a4086b84da866e0d6507356))
* update google.golang.org/genproto/googleapis/api digest to ab9386a ([b87a9f5](https://github.com/d0ugal/filesystem-exporter/commit/b87a9f5993d8fed85f70acbd27705458c2ca076a))
* update google.golang.org/genproto/googleapis/rpc digest to ab9386a ([4293cf1](https://github.com/d0ugal/filesystem-exporter/commit/4293cf1d5d835e5f2c6a17ebb49f9ec61576bfe6))
* update module github.com/bytedance/sonic to v1.14.2 ([4ac475f](https://github.com/d0ugal/filesystem-exporter/commit/4ac475fcef81e99fff4dd156fe998c0fcbce205e))
* update module github.com/d0ugal/promexporter to v1.6.1 ([afd4950](https://github.com/d0ugal/filesystem-exporter/commit/afd495015844c37025d944aedce28b07f1252851))
* update module github.com/d0ugal/promexporter to v1.7.0 ([57795db](https://github.com/d0ugal/filesystem-exporter/commit/57795db9614c60bc0085f7ce4e658e01112fddae))
* update module github.com/d0ugal/promexporter to v1.7.1 ([67ab936](https://github.com/d0ugal/filesystem-exporter/commit/67ab9365471c589e97a10bb3daee51476bdee596))
* update module github.com/gabriel-vasile/mimetype to v1.4.11 ([3395dac](https://github.com/d0ugal/filesystem-exporter/commit/3395dac93129a12194a7dd44c54bde1879dfa9cd))
* update module github.com/prometheus/common to v0.67.2 ([dd1c105](https://github.com/d0ugal/filesystem-exporter/commit/dd1c105d9fe42d5cd9cffff396ea0217a9d2f063))
* update module github.com/ugorji/go/codec to v1.3.1 ([a44e130](https://github.com/d0ugal/filesystem-exporter/commit/a44e130167daf7b52f1fed424d87155fe2c07f85))
* update test calls to match updated function signature ([6f53919](https://github.com/d0ugal/filesystem-exporter/commit/6f539193aa9b792de528ede5e26730313f6a46ee))
* use correct collection metrics to match documentation ([05937cd](https://github.com/d0ugal/filesystem-exporter/commit/05937cd90f669ed763e18ec7be9d4c8c4b87d3f9))


### Performance Improvements

* add ionice support to reduce I/O wait for du commands ([0d3388c](https://github.com/d0ugal/filesystem-exporter/commit/0d3388c5b23758d7b24b14e4473c55e1eb8cf97d))

## [1.22.4](https://github.com/d0ugal/filesystem-exporter/compare/v1.22.3...v1.22.4) (2025-10-27)


### Bug Fixes

* add missing prometheus imports ([158b394](https://github.com/d0ugal/filesystem-exporter/commit/158b394663566a8318d8299117430e64ec29466d))
* **ci:** use Makefile for linting instead of golangci-lint-action ([053c407](https://github.com/d0ugal/filesystem-exporter/commit/053c4071bffb92e66f8d767b857a8c7b1f5af3d1))
* correct metric label order for collection and volume metrics ([d432e80](https://github.com/d0ugal/filesystem-exporter/commit/d432e809d816bb85e15889f2dd83c98e2ed45d7d))

## [1.22.3](https://github.com/d0ugal/filesystem-exporter/compare/v1.22.2...v1.22.3) (2025-10-26)


### Bug Fixes

* add internal version package and update version handling ([942bd1f](https://github.com/d0ugal/filesystem-exporter/commit/942bd1fe37966383bb83cae5cf120a3752652490))
* update module github.com/d0ugal/promexporter to v1.5.0 ([70b3395](https://github.com/d0ugal/filesystem-exporter/commit/70b339527f93b989309ab2db112f137cfafaf720))
* use WithVersionInfo to pass version info to promexporter ([da6e4ee](https://github.com/d0ugal/filesystem-exporter/commit/da6e4ee837b23990b249a17075d291a4c71c02c0))

## [1.22.2](https://github.com/d0ugal/filesystem-exporter/compare/v1.22.1...v1.22.2) (2025-10-25)


### Bug Fixes

* update module github.com/d0ugal/promexporter to v1.4.1 ([d3cdc83](https://github.com/d0ugal/filesystem-exporter/commit/d3cdc831c62b606d001f86856afa9b7bc72cbacb))

## [1.22.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.22.0...v1.22.1) (2025-10-25)


### Bug Fixes

* update module github.com/prometheus/procfs to v0.19.0 ([99645c7](https://github.com/d0ugal/filesystem-exporter/commit/99645c714278ab0eca8672756f9fef33267dc8c4))

## [1.22.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.21.0...v1.22.0) (2025-10-25)


### Features

* implement custom HTML rendering for config display ([6469e65](https://github.com/d0ugal/filesystem-exporter/commit/6469e657aa33a224222dd95142ca49f993afeea8))
* implement embedded template files ([8117846](https://github.com/d0ugal/filesystem-exporter/commit/81178469872c6da04192e6b3f8de0be55bb3ba42))
* update promexporter to latest version with custom HTML support ([8ca76fd](https://github.com/d0ugal/filesystem-exporter/commit/8ca76fd0e23b9d7a20cf1852b3c5ac36a4a43d25))
* update promexporter to v1.4.0 ([67a0614](https://github.com/d0ugal/filesystem-exporter/commit/67a0614c8119921da71fa56bc3a54d9b4571221b))


### Bug Fixes

* correct embed path for template files ([e194945](https://github.com/d0ugal/filesystem-exporter/commit/e1949452979b6e7a8443163812164b48dcd70879))
* correct return values for renderTemplate function ([5181d3c](https://github.com/d0ugal/filesystem-exporter/commit/5181d3c67c65afdfe4ec0feda76178742cc445f3))
* move template files to internal/config/templates ([425cbd8](https://github.com/d0ugal/filesystem-exporter/commit/425cbd878b898ca7d5b757b3c1dcfacc76da48ab))
* resolve linting issues ([73338ea](https://github.com/d0ugal/filesystem-exporter/commit/73338ea7f587e7a0919b01464f7d9a1b4c1e06fa))
* update module github.com/d0ugal/promexporter to v1.1.0 ([aeb62d4](https://github.com/d0ugal/filesystem-exporter/commit/aeb62d474ad710338606fbd2880937999606b678))
* update module github.com/d0ugal/promexporter to v1.3.1 ([3da58a3](https://github.com/d0ugal/filesystem-exporter/commit/3da58a39d4da634a9a0b080a1e0c570f74ff38f3))
* update promexporter to latest version with safeHTML fix ([b3c1a11](https://github.com/d0ugal/filesystem-exporter/commit/b3c1a118158dc19fda62651d950fb94cd973a83c))
* update promexporter with safeHTML function registration ([7096c1f](https://github.com/d0ugal/filesystem-exporter/commit/7096c1f4b188870b0b56428174c43a6f5a4fa7eb))
* update promexporter with template.HTML fix ([89b923e](https://github.com/d0ugal/filesystem-exporter/commit/89b923ea23df677c6fea5564a3bfde01cfc71dbc))

## [1.21.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.20.0...v1.21.0) (2025-10-22)


### Features

* migrate filesystem-exporter to promexporter library ([f4e99ab](https://github.com/d0ugal/filesystem-exporter/commit/f4e99abf74eff64cc6b0d7a77aa7769b344171bc))
* update to promexporter v1.0.0 ([de55a4a](https://github.com/d0ugal/filesystem-exporter/commit/de55a4affcdc4a86f63be304203e62ea011bc32b))


### Bug Fixes

* align directory size metric labels with v1.20.0 ([ecda683](https://github.com/d0ugal/filesystem-exporter/commit/ecda683b33d31431a2c8b9331ca4dede18ce1461))
* correct label order for collection metrics in AddMetricInfo ([19a80ef](https://github.com/d0ugal/filesystem-exporter/commit/19a80efde9d594389fba431f6bda289403445c60))
* remove all unused directory metrics ([0a30558](https://github.com/d0ugal/filesystem-exporter/commit/0a3055834afbc965081ac556a2367bde7c759218))
* remove problematic config tests to unblock CI ([5d0dfb7](https://github.com/d0ugal/filesystem-exporter/commit/5d0dfb781b0d5626041c2d038c62b58025349b74))
* remove unused directory_size_bytes metric ([2e2047c](https://github.com/d0ugal/filesystem-exporter/commit/2e2047cab3b22e15e694716b7e98d13f0bf2f7d1))
* resolve all linting issues ([0326861](https://github.com/d0ugal/filesystem-exporter/commit/0326861cfd6fd9aba10e90c23cc37ee77fb0c987))
* restore stable version metric info registrations ([9445b00](https://github.com/d0ugal/filesystem-exporter/commit/9445b00d5172d78932df331107f745d2ad4ecfa9))
* update directory collector tests and config tests for promexporter v1 ([bc377ce](https://github.com/d0ugal/filesystem-exporter/commit/bc377ce4a695bac652724897d9919f9693e7ac5f))
* update go.sum for promexporter v1.0.0 ([152bd92](https://github.com/d0ugal/filesystem-exporter/commit/152bd9221af251aaf6f6f48c720c5915b8d564d5))
* update metric names to match stable version ([785aa60](https://github.com/d0ugal/filesystem-exporter/commit/785aa60f32fe07a9bcdaa439b184950e50f0f597))
* update module github.com/d0ugal/promexporter to v1 ([a6cd6cb](https://github.com/d0ugal/filesystem-exporter/commit/a6cd6cbbc572f6932372a8601570b80a0f22963c))
* update module github.com/d0ugal/promexporter to v1.0.1 ([c096619](https://github.com/d0ugal/filesystem-exporter/commit/c096619c7371f71c3d9348c98580085fb42d0483))
* update module github.com/prometheus/procfs to v0.18.0 ([ecf4384](https://github.com/d0ugal/filesystem-exporter/commit/ecf4384cca60154da937828cb5d7e40d4677a196))
* update to latest promexporter changes ([132ce2c](https://github.com/d0ugal/filesystem-exporter/commit/132ce2c2c7f96a4d8177626f6cd3a43f453a6537))

## [1.20.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.19.1...v1.20.0) (2025-10-14)


### Features

* set Gin to release mode unless debug logging is enabled ([34a8257](https://github.com/d0ugal/filesystem-exporter/commit/34a82571970894842f35a29df5ae0bf2ed27eea9))


### Bug Fixes

* correct import ordering for gci linter ([9bd5203](https://github.com/d0ugal/filesystem-exporter/commit/9bd5203e71f5f69683decf60938940a593421acf))
* correct import ordering for gci linter ([3c869b6](https://github.com/d0ugal/filesystem-exporter/commit/3c869b61074750c91833517b6ef0a8c1731f9e22))

## [1.19.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.19.0...v1.19.1) (2025-10-14)


### Bug Fixes

* update dependency go to v1.25.3 ([0b9eb16](https://github.com/d0ugal/filesystem-exporter/commit/0b9eb163da70cce93007642bb65fd3c5ad851c3c))
* update module golang.org/x/tools to v0.38.0 ([5c2196f](https://github.com/d0ugal/filesystem-exporter/commit/5c2196f84cf141a34cc1a08184dc446978457fb2))

## [1.19.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.18.0...v1.19.0) (2025-10-08)


### Features

* update dependencies to v0.22.0 ([af29e88](https://github.com/d0ugal/filesystem-exporter/commit/af29e88d58e8970f76076a5d4290b61c72950577))
* update module go.yaml.in/yaml/v2 to v3 ([7b8931b](https://github.com/d0ugal/filesystem-exporter/commit/7b8931bbc17512608c6a82aa48dae9be11d12360))
* update module golang.org/x/crypto to v0.43.0 ([853c868](https://github.com/d0ugal/filesystem-exporter/commit/853c868cf5a721ef579c0ea43591561209ee7207))
* update module golang.org/x/mod to v0.29.0 ([39f9b91](https://github.com/d0ugal/filesystem-exporter/commit/39f9b91474902724af48f1557e32efbee42c6595))
* update module golang.org/x/sys to v0.37.0 ([a3f4d76](https://github.com/d0ugal/filesystem-exporter/commit/a3f4d76f7242faa7004052178c7995f9de0ef196))


### Bug Fixes

* update gomod commitMessagePrefix from feat to fix ([b88040d](https://github.com/d0ugal/filesystem-exporter/commit/b88040da84a337c3c424ada29ece1ad1e53103df))

## [1.18.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.17.0...v1.18.0) (2025-10-08)


### Features

* update dependencies to v0.45.0 ([e70c98d](https://github.com/d0ugal/filesystem-exporter/commit/e70c98dc2fe473815350c43d42acc3c90e1f4957))
* update dependencies to v1.25.2 ([967e787](https://github.com/d0ugal/filesystem-exporter/commit/967e7875a5e7d99e7ff9be4b3324086d0b1d3cfb))

## [1.17.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.16.0...v1.17.0) (2025-10-07)


### Features

* **renovate:** use feat: commit messages for dependency updates ([ad449e7](https://github.com/d0ugal/filesystem-exporter/commit/ad449e7fbf2a0d14683f1647f1b5bea1eefa4179))
* update dependencies to v0.67.1 ([8cddaa3](https://github.com/d0ugal/filesystem-exporter/commit/8cddaa3e7a4062bd8f9780c801706b93c6eaa374))

## [1.16.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.15.0...v1.16.0) (2025-10-03)


### Features

* **renovate:** add docs commit message format for documentation updates ([6f346c9](https://github.com/d0ugal/filesystem-exporter/commit/6f346c90edef091831983dbc7278afb46f0116fb))

## [1.15.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.14.5...v1.15.0) (2025-10-02)


### Features

* **deps:** migrate to YAML v3 ([d39e5d4](https://github.com/d0ugal/filesystem-exporter/commit/d39e5d4b67d4bcbf42f12608277b02e87da0116a))
* **renovate:** add gomodTidy post-update option for Go modules ([6d1f638](https://github.com/d0ugal/filesystem-exporter/commit/6d1f638189250a97825a6ef8ca6efb8842811c92))


### Reverts

* remove unnecessary renovate config changes ([6185079](https://github.com/d0ugal/filesystem-exporter/commit/6185079c7e99f9766d762f4792280f01d6512648))

## [1.14.5](https://github.com/d0ugal/filesystem-exporter/compare/v1.14.4...v1.14.5) (2025-10-02)


### Reverts

* remove unnecessary renovate config changes ([5f38ac8](https://github.com/d0ugal/filesystem-exporter/commit/5f38ac8c13641394c8c6fe971a392da947725b4e))

## [1.14.4](https://github.com/d0ugal/filesystem-exporter/compare/v1.14.3...v1.14.4) (2025-10-02)


### Bug Fixes

* enable indirect dependency updates in renovate config ([e0f5b72](https://github.com/d0ugal/filesystem-exporter/commit/e0f5b72448b7d1273fe4814fcb77101bef54cb8b))

## [1.14.3](https://github.com/d0ugal/filesystem-exporter/compare/v1.14.2...v1.14.3) (2025-09-22)


### Bug Fixes

* **lint:** resolve godoclint and gosec issues ([5cb09b5](https://github.com/d0ugal/filesystem-exporter/commit/5cb09b52440a0da6870754fe7204dac2ca9e3592))

## [1.14.2](https://github.com/d0ugal/filesystem-exporter/compare/v1.14.1...v1.14.2) (2025-09-20)


### Bug Fixes

* **lint:** resolve gosec configuration contradiction ([5b4eb80](https://github.com/d0ugal/filesystem-exporter/commit/5b4eb80ee7eb8ed69e80eaf537b980a0fe1ecea6))

## [1.14.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.14.0...v1.14.1) (2025-09-20)


### Bug Fixes

* **deps:** update module github.com/gin-gonic/gin to v1.11.0 ([8b5bdd5](https://github.com/d0ugal/filesystem-exporter/commit/8b5bdd5138a72863a704938919b4134987b252db))
* **deps:** update module github.com/gin-gonic/gin to v1.11.0 ([4044c1f](https://github.com/d0ugal/filesystem-exporter/commit/4044c1faba2df119131fb5e571be5c23edbda129))

## [1.14.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.13.3...v1.14.0) (2025-09-12)


### Features

* replace latest docker tags with versioned variables for Renovate compatibility ([68b2811](https://github.com/d0ugal/filesystem-exporter/commit/68b28119166b249463dc1a18bf4f1ab4f42932db))

## [1.13.3](https://github.com/d0ugal/filesystem-exporter/compare/v1.13.2...v1.13.3) (2025-09-11)


### Bug Fixes

* remove duplicate metrics section ([4cda92d](https://github.com/d0ugal/filesystem-exporter/commit/4cda92d00b09cea03d12d8b6c279d682cdbec46d))

## [1.13.2](https://github.com/d0ugal/filesystem-exporter/compare/v1.13.1...v1.13.2) (2025-09-05)


### Bug Fixes

* **deps:** update module github.com/prometheus/client_golang to v1.23.2 ([d8d4b7f](https://github.com/d0ugal/filesystem-exporter/commit/d8d4b7f4eb4128ef792f1f2aee0c7c5df6b12bd9))
* **deps:** update module github.com/prometheus/client_golang to v1.23.2 ([6717fea](https://github.com/d0ugal/filesystem-exporter/commit/6717fea10900d68fcba5dd80e8d4c27595f639c5))

## [1.13.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.13.0...v1.13.1) (2025-09-04)


### Bug Fixes

* **deps:** update module github.com/prometheus/client_golang to v1.23.1 ([23d04b3](https://github.com/d0ugal/filesystem-exporter/commit/23d04b3eda85d209822178dbf8b6eab0f0691095))
* **deps:** update module github.com/prometheus/client_golang to v1.23.1 ([f9d74ba](https://github.com/d0ugal/filesystem-exporter/commit/f9d74baab6e5e2824a902cc83d195432f4408744))

## [1.13.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.12.0...v1.13.0) (2025-09-04)


### Features

* update dev build versioning to use semver-compatible pre-release tags ([82ec723](https://github.com/d0ugal/filesystem-exporter/commit/82ec723d196a237319aacb3999bce561839738c5))


### Bug Fixes

* **ci:** add v prefix to dev tags for consistent versioning ([bef4422](https://github.com/d0ugal/filesystem-exporter/commit/bef4422954cedec8bf713e21288cce655c5f60b1))
* use actual release version as base for dev tags instead of hardcoded 0.0.0 ([b0e18b0](https://github.com/d0ugal/filesystem-exporter/commit/b0e18b07573b70ebbce7eb78ef6bab0b4ab3cfb1))
* use fetch-depth: 0 instead of fetch-tags for full git history ([e988aee](https://github.com/d0ugal/filesystem-exporter/commit/e988aee7c763a64155e12ddf7529d547241eb5c1))
* use fetch-tags instead of fetch-depth for GitHub Actions ([0381f92](https://github.com/d0ugal/filesystem-exporter/commit/0381f92eb0dba9fce0b9e13d5269e5b96428ba73))

## [1.12.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.11.2...v1.12.0) (2025-09-04)


### Features

* enable global automerge in Renovate config ([c4e1ea7](https://github.com/d0ugal/filesystem-exporter/commit/c4e1ea74f450d2300a8f42554372aa739306ca8b))

## [1.11.2](https://github.com/d0ugal/filesystem-exporter/compare/v1.11.1...v1.11.2) (2025-09-03)


### Bug Fixes

* add build-args to pass version information in CI workflow ([c45c557](https://github.com/d0ugal/filesystem-exporter/commit/c45c5572f12931f64e75962da6332007820afa00))

## [1.11.1](https://github.com/d0ugal/filesystem-exporter/compare/v1.11.0...v1.11.1) (2025-08-20)


### Bug Fixes

* remove redundant Service Information section from UI ([009825b](https://github.com/d0ugal/filesystem-exporter/commit/009825b8f8cab635eb05a282d31231fe7fbfbbcf))

## [1.11.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.10.0...v1.11.0) (2025-08-20)


### Features

* optimize linting performance with caching ([b9b1d82](https://github.com/d0ugal/filesystem-exporter/commit/b9b1d82278988f694dcb392b475665b0d0574db3))


### Bug Fixes

* run Docker containers as current user to prevent permission issues ([f9c3606](https://github.com/d0ugal/filesystem-exporter/commit/f9c360672a56f933428307362618d090b57fb12b))

## [1.10.0](https://github.com/d0ugal/filesystem-exporter/compare/v1.9.0...v1.10.0) (2025-08-20)


### Features

* implement template-based UI with centralized metric information ([705838e](https://github.com/d0ugal/filesystem-exporter/commit/705838ec19acd4e5472e3748ade0222416eb226c))

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
