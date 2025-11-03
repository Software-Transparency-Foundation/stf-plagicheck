# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-10-28

### Added
- Initial public release of Plagicheck
- Winnowing fingerprint (WFP) generation for source code files
- Full file matching against STF Open Knowledge Base
- Code snippet matching with line range identification
- Multi-threaded parallel processing support (-T flag)
- Debug mode for detailed processing information (-d flag)
- Support for scanning individual files, directories, and WFP files
- JSON output format for easy integration
- Configurable hit threshold for snippet matching (--min-hits flag)
- Progress bar display for long-running operations
- Command-line interface with comprehensive options
- Complete test suite with unit and integration tests
- Comprehensive documentation (README, CONTRIBUTING, LICENSE)

### Features
- Fast and efficient code scanning using winnowing algorithm
- Range merging to consolidate nearby matches
- Automatic filtering of single-line matches
- Support for multiple file types (source code only)
- Integration with SCANOSS LDB knowledge base
- Integration with SCANOSS snippet scanning engine

### Documentation
- Complete README with installation and usage instructions
- Contributing guidelines
-GPL-2.0-only license
- Code examples and test files
- Makefile with build automation

## [Unreleased]

### Planned
- Additional output formats (CSV, SARIF)
- Configuration file support
- Incremental scanning support
- Performance optimizations for large codebases

