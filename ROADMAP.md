# Parts Development Roadmap

This document tracks planned improvements and features for the Parts project. Items are organized by priority and category to help guide development efforts.

## Recently Completed ✅

- [x] Add proper error handling instead of `panic()` calls
- [x] Add command-line argument validation and help text
- [x] Create idempotent editing tests to prevent whitespace growth
- [x] Implement comprehensive Makefile with development tools
- [x] Add project documentation (README, ROADMAP)
- [x] Migrate to standard Go testing framework (`testing` package)
- [x] Add dry-run mode (`--dry-run` flag) to preview changes
- [x] Create proper CLI interface with flags library (cobra/flag)
- [x] Support multiple comment character styles (e.g., `//`, `/*`, `--`)
- [x] Enhanced comment markers with proper header/footer blocks for visual clarity
- [x] Add `--remove` flag to cleanly remove partials sections from files
- [x] Refactor large monolithic file into focused modules with clear separation of concerns
- [x] Clarify comment styles documentation to explain they're for language-appropriate delimiters
- [x] Convert comment style examples to table format for better readability

---

## Critical - Bug Fixes 🚨

These issues affect reliability and should be fixed immediately.

### Deprecated Go APIs
- [x] Replace `ioutil.ReadDir()` with `os.ReadDir()` in `src/build.go`
- [x] Replace `ioutil.WriteFile()` with `os.WriteFile()` in `src/build.go` and `src/remove.go`
- [x] Remove `ioutil` import entirely (deprecated since Go 1.16)

### Error Handling Bugs
- [x] Fix silent error in `ExpandTildePrefix()` - now returns error instead of ignoring
- [x] Propagate tilde expansion errors up through `NewPartialsBuildCommand()` and `NewPartialsRemoveCommand()`

### File Permission Bug
- [x] Preserve original file permissions instead of hardcoding `0600`
- [x] Read original mode with `os.Stat()` before modification, restore after write

### Inconsistent Behavior
- [x] Add success message to build command (remove command has one, build doesn't)
- [ ] Standardize error message formatting (capitalization, punctuation, `%w` for wrapping)
- [ ] Success/status messages should go to stderr, not stdout (interferes with piping)
- [ ] Remove command doesn't show removed character count in non-dry-run mode (only dry-run shows it)

---

## High Priority 🔥

### Code Quality & Architecture
- [x] Extract magic strings to constants (`PARTIALS>>>>>`, `PARTIALS<<<<<`, separator line)
- [ ] Extract shared code from `build.go` and `remove.go` (~100 lines duplicated):
  - `GetStartFlag()` / `GetEndFlag()` methods (identical implementations)
  - `SetDryRun()` method
  - Comment style resolution logic
  - Dry-run handling pattern
  - File permission preservation logic (now duplicated in both)
- [ ] Implement structured logging with verbosity levels (`-v`, `-vv`)
- [ ] Add input validation:
  - Verify paths are files/directories as expected
  - Check for symlink traversal attacks
  - Validate aggregate file is writable
  - Warn if partials directory is empty

### Go Version & CI
- [ ] Update `go.mod` from Go 1.17 to Go 1.21 (minimum supported)
- [ ] Update CI matrix to test Go 1.21, 1.22, 1.23
- [ ] Update CONTRIBUTING.md Go version requirement

### Essential Features
- [ ] Add backup functionality (create `.bak` files before modification)
- [ ] Add `--verbose` flag to show files being processed
- [ ] Add `--quiet` flag to suppress success messages (useful for scripts/cron)
- [x] Add configuration file support (`.parts.yaml` manifest) with `apply`, `remove`, `sync`, `init` subcommands
- [x] Add `own` mode for fully-managed files (no markers)
- [x] Add `sync` subcommand for reverse-syncing target changes back to partials

---

## Medium Priority 📋

### Testing Improvements
- [ ] Add edge case tests:
  - Empty aggregate file
  - Partials directory with subdirectories (currently silently ignored)
  - Files with no trailing newline
  - Files with mixed line endings (CRLF vs LF)
  - Very large files (memory efficiency)
- [ ] Add error path tests:
  - Permission denied on write
  - Disk full scenario
  - File modified during processing
- [ ] Add integration tests with real SSH config scenarios
- [ ] Add benchmarks for performance testing
- [ ] Implement code coverage reporting in CI
- [ ] Suppress stdout in tests (success messages clutter test output)
- [ ] Refactor CLI tests to avoid recreating cobra commands manually (fragile pattern in `cmd/*_test.go`)

### Features & Functionality
- [ ] Implement watch mode to auto-rebuild on partial file changes
- [ ] Add support for nested partial directories
- [ ] Add merge conflict detection and resolution
- [ ] Support `~username/path` expansion (other user's home directory)
- [ ] Improve auto-detection warnings (log detected style, warn on unknown extensions)

### Documentation & Usability
- [ ] Add troubleshooting guide with common issues
- [ ] Add man page generation
- [ ] Create usage examples for different file types (not just SSH)
- [ ] Document security considerations
- [ ] Add changelog and versioning
- [ ] Add limitations section to README

---

## Low Priority 📅

### Distribution & Packaging
- [ ] Create release automation with goreleaser
- [ ] Package for Homebrew
- [ ] Add shell completion scripts (bash, zsh, fish)
- [ ] Generate checksums for releases
- [ ] Create Docker container for isolated usage

### Performance & Scalability
- [ ] Optimize for large files (streaming processing)
- [ ] Add concurrent processing for multiple files
- [ ] Implement caching for unchanged partials
- [ ] Add file locking for concurrent access safety

### Security & Reliability
- [ ] Implement atomic file operations (write to temp, rename)
- [ ] Add rollback functionality on failure
- [ ] Add file integrity checking (checksums)

### Code Quality
- [ ] Standardize receiver types (pointer vs value) across all methods
- [ ] Add Makefile targets: `test-watch`, `coverage`, `deps-check`, `security`
- [ ] Extract CLI flag setup to reduce test duplication in `cmd/root_test.go`

---

## Future/Advanced Features 🚀

- [ ] Add plugin system for custom processors
- [ ] Implement file format validation (SSH config syntax checking)
- [ ] Add support for encrypted partials
- [ ] Create web UI for managing partials
- [ ] Add integration with version control systems
- [ ] Support for remote partial sources (HTTP/S3)

---

## Contributing

When working on items from this roadmap:

1. **Start with Critical items** - these are bugs affecting reliability
2. **Create issues** for larger features to discuss implementation
3. **Update this file** when items are completed
4. **Add new items** as they're identified during development

## Versioning Strategy

- **v0.x**: Current phase, stabilizing core functionality
- **v1.0**: First stable release after Critical bugs fixed
- **v2.0+**: Advanced features and major enhancements

## Breaking Changes Log

Track API changes that may affect users importing this as a library:

- `NewPartialsBuildCommand()` now returns `(PartialsBuildCommand, error)` instead of `PartialsBuildCommand`
- `NewPartialsRemoveCommand()` now returns `(PartialsRemoveCommand, error)` instead of `PartialsRemoveCommand`
- `ExpandTildePrefix()` now returns `(string, error)` instead of `string`
- Added `MustExpandTildePrefix()` for cases where panic on error is acceptable

---

*Last updated: 2026-02-16*
