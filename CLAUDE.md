# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Parts is a Go utility for managing dotfiles and configuration files from partial fragments. It supports two modes of operation:

1. **Legacy CLI** — positional-argument interface for merging partials into a single aggregate file using comment markers
2. **Manifest mode** — `.parts.yaml`-driven multi-target management with `apply`, `remove`, `sync`, and `init` subcommands

## Commands

### Building
- `make parts` - Build the main parts binary to `bin/parts`
- `go build -o bin/parts ./main.go` - Direct Go build command

### Testing
- `make tests` - Build and run all tests
- `go build -o bin/tests tests/*.go && bin/tests` - Direct test build and run

### Legacy Usage (positional args)
- `bin/parts ~/.ssh/config ~/.ssh/config.d "#"` - Merge SSH config partials
- `bin/parts app.js ./partials "//"` - Merge JavaScript partials with slash comments
- `bin/parts styles.css ./css-partials "/*"` - Merge CSS partials with block comments
- `bin/parts config.py ./python-configs "auto"` - Auto-detect comment style from file extension
- `bin/parts --dry-run ~/.ssh/config ~/.ssh/config.d "#"` - Preview changes without modifying files
- `bin/parts -n ~/.ssh/config ~/.ssh/config.d "#"` - Preview changes (short flag)
- `bin/parts --remove ~/.ssh/config "#"` - Remove partials section from file
- `bin/parts --help` - Show help and usage information
- `make ssh` - Quick SSH config merge using default paths

### Manifest Mode (subcommands)
- `bin/parts init` - Generate a skeleton `.parts.yaml` manifest
- `bin/parts apply` - Apply all manifest targets
- `bin/parts apply ssh` - Apply only the 'ssh' target
- `bin/parts apply --dry-run` - Preview changes without modifying files
- `bin/parts remove` - Remove managed content from all targets (merge: strip markers, own: delete file)
- `bin/parts remove ssh` - Remove only the 'ssh' target
- `bin/parts sync` - Sync changes from target files back into partials
- `bin/parts sync --dry-run` - Preview what would be synced

## Architecture

### Core Components
- **main.go**: Entry point that calls the cobra CLI framework
- **cmd/root.go**: Cobra CLI root command (legacy positional-arg interface) and subcommand registration
- **cmd/apply.go**: `apply` subcommand — orchestrates manifest targets
- **cmd/remove_manifest.go**: `remove` subcommand — removes managed content from targets
- **cmd/sync.go**: `sync` subcommand — reverse-syncs target changes back to partials
- **cmd/init.go**: `init` subcommand — generates skeleton `.parts.yaml`
- **src/build.go**: Build/merge engine for partial files into aggregate file
- **src/remove.go**: Remove engine for cleaning partials sections from files
- **src/own.go**: Own mode engine for whole-file management (no markers)
- **src/manifest.go**: `.parts.yaml` manifest parsing, validation, and defaults resolution
- **src/sync.go**: Sync engine — extracts sections from targets and writes back to partials
- **src/comments.go**: Comment style detection and resolution
- **src/constants.go**: Shared constants (marker strings)
- **src/utils.go**: Utility functions (path expansion, etc.)

### Key Functionality

**Merge mode** (default — legacy CLI and manifest `mode: merge`):
1. Reading an aggregate file and removing any existing managed section (between PARTIALS>>>>> and PARTIALS<<<<< markers)
2. Reading all files from the partials directory
3. Appending the partial file contents between the comment markers
4. Writing the updated aggregate file

**Own mode** (manifest `mode: own`):
1. Reading all files from the partials directory
2. Concatenating them (with optional source comments) into a single output
3. Writing the entire target file from the concatenated partials

**Sync** (manifest `sync` subcommand):
1. Reading the target file and extracting sections by `# Source: <path>` comments
2. Comparing extracted content against the original partial files
3. Writing changed content back to the partial source files

### File Structure
```
main.go                     # Entry point
cmd/
  ├── root.go                # Legacy CLI + subcommand registration
  ├── root_test.go           # Legacy CLI tests
  ├── remove_test.go         # Legacy remove flag tests
  ├── apply.go               # 'apply' subcommand
  ├── apply_test.go          # Apply subcommand tests
  ├── remove_manifest.go     # 'remove' subcommand (manifest mode)
  ├── remove_manifest_test.go
  ├── sync.go                # 'sync' subcommand
  ├── sync_test.go
  ├── init.go                # 'init' subcommand
  └── init_test.go
src/
  ├── build.go               # Merge engine
  ├── remove.go              # Remove engine
  ├── own.go                 # Own mode engine
  ├── manifest.go            # .parts.yaml parsing & validation
  ├── sync.go                # Sync engine
  ├── comments.go            # Comment style detection
  ├── constants.go           # Shared marker constants
  ├── utils.go               # Utility functions
  ├── partialsBuildCommand_test.go
  ├── commentStyles_test.go
  ├── manifest_test.go
  ├── own_test.go
  ├── sync_test.go
  └── utils_test.go
```

## Development Notes

### Go Version
- Uses Go 1.17 (specified in go.mod)

### Testing Pattern
- Uses standard Go testing framework (`testing` package)
- Test files are located in `src/` and `cmd/` directories with `*_test.go` naming
- Tests use `t.TempDir()` for temporary directories 
- Comprehensive test coverage including CLI functionality, error handling and idempotency
- Run tests with `make test` or `go test ./...` to run all tests
- CLI tests in `cmd/root_test.go` test argument validation, help output, and dry-run functionality

## Development Planning

For project roadmap, feature planning, and improvement tracking, see:
- **[ROADMAP.md](./ROADMAP.md)** - Comprehensive development roadmap with prioritized features and improvements

### Quick Development Notes
- Focus on high-priority items first (marked 🔥 in roadmap)
- Always add tests for new features
- Update README.md when adding user-facing features
- Use proper error handling patterns established in the codebase