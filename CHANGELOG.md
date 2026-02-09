# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [0.1.2][] - 2026-02-10

### Changed

* PNG encoding and decoding functions have been sped up by switching
  to the `github.com/woozymasta/png` library.

### Fixed

* An over-check in the `convert` command prevented conversion to
  non-BCn-based images.

[0.1.2]: https://github.com/WoozyMasta/imageset-packer/compare/v0.1.1...v0.1.2

## [0.1.1][] - 2026-02-07

### Added

* Added output format selection for atlas/convert flows:
  `bgra8` (default) and new `dxt1`, `dxt5`.
* Added encoder quality option (`0..10`) for DXT output
  (`0` = library default).

### Changed

* Migrated internal DDS/EDDS/DXT paths to external packages:
  `github.com/woozymasta/bcn` and `github.com/woozymasta/edds`.
* Encoding/decoding performance is significantly improved by BCn-side
  parallel processing (workers auto-scaled via `GOMAXPROCS` by default).

[0.1.1]: https://github.com/WoozyMasta/imageset-packer/compare/v0.1.0...v0.1.1

## [0.1.0][] - 2026-01-25

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/imageset-packer/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
