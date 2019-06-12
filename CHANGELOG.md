# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
[Unreleased]: https://github.com/mec07/rununtil/compare/v0.1.0...HEAD

## [0.1.0] - 2019-06-12
[0.1.0]: https://github.com/mec07/rununtil/compare/v0.0.1...v0.1.0
### Changed
- There are now RunnerFunc and ShutdownFunc function types to clarify the usage of this library (backwards incompatible change)

### Added
- Implemented the rununtil.Killed method, which allows you to test a function that uses rununtil.KillSignal


## [0.0.1] - 2019-06-05
[0.0.1]: https://github.com/mec07/rununtil/releases/tag/v0.0.1
### Added
- initial commit of rununtil library


