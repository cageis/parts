# Manifest-Driven Dotfile Management

## Problem

Parts currently operates on a single target at a time via positional CLI args. Users managing multiple dotfiles must invoke Parts repeatedly. There's no way to:
- Manage multiple targets declaratively from a single config
- Detect changes in target files and sync them back to the partials repo
- Own entire files (not just merged sections)

## Design

### CLI Interface

Existing CLI is preserved exactly as-is (manifestless fallback):
```bash
parts ~/.ssh/config ~/.ssh/config.d "#"
parts --dry-run ~/.ssh/config ~/.ssh/config.d "#"
parts --remove ~/.ssh/config "#"
```

New manifest-driven subcommands:
```bash
parts apply                    # Apply all targets in .parts.yaml
parts apply --dry-run          # Preview all targets
parts apply ssh                # Apply only the "ssh" target by name
parts remove                   # Remove managed sections from all targets
parts remove vim               # Remove only the "vim" target
parts sync                     # Detect changes in targets, pull back into partials
parts sync ssh                 # Sync only the "ssh" target
parts init                     # Generate a skeleton .parts.yaml
```

Resolution: subcommands (apply/remove/sync/init) vs positional args distinguish manifest mode from legacy mode. No ambiguity.

### Manifest Format (.parts.yaml)

```yaml
defaults:
  comment: "auto"
  backup: true

targets:
  ssh:
    target: ~/.ssh/config
    partials: ./ssh/config.d/
    comment: "#"
    mode: merge

  vimrc:
    target: ~/.vimrc
    partials: ./vim/
    mode: own

  gitconfig:
    target: ~/.gitconfig
    partials: ./git/
    comment: "#"
    mode: merge
```

### Modes

- **`merge`**: Uses existing marker-based mechanism. Target file can have user content outside the PARTIALS markers. Comment style required.
- **`own`**: Parts writes the entire file from concatenated partials. No markers. Comment style optional (for source-path comments only).

### Subcommand Behavior

**`parts apply [target-name...]`**
- Reads `.parts.yaml` from current directory
- For each target (or named subset):
  - `merge` mode: runs existing PartialsBuildCommand logic
  - `own` mode: concatenates partials, writes entire file
- Supports `--dry-run`
- Creates `.bak` backup if `backup: true`

**`parts remove [target-name...]`**
- `merge` mode: runs existing PartialsRemoveCommand logic (strips markers + content)
- `own` mode: deletes the target file (or restores from .bak if available)
- Supports `--dry-run`

**`parts sync [target-name...]`**
- For `merge` targets: extracts content between PARTIALS markers from the target file, diffs against current partials, writes changes back to partial files
- For `own` targets: diffs entire target file against concatenated partials, writes back
- Challenge: mapping changed content back to individual partial files. Strategy:
  - Source comments (`# Source: ./ssh/config.d/work.conf`) already exist in merged output
  - Use these markers to split the managed section back into per-partial chunks
  - Write each chunk back to its source file
- Supports `--dry-run` to preview what would change

**`parts init`**
- Generates a skeleton `.parts.yaml` with commented examples
- If run in a directory with existing partial-like structure, attempts to auto-detect targets

### Defaults Inheritance

Per-target values override `defaults`. Missing values fall back to defaults. If no default and no per-target value, use sensible built-in defaults:
- `comment`: `"auto"` (detect from file extension)
- `mode`: `"merge"` (safer default, preserves existing file content)
- `backup`: `false`

### Error Handling

- Missing `.parts.yaml`: clear error message pointing to `parts init`
- Invalid target name in selective apply: error listing available targets
- Missing partials directory: error per-target, continue others
- Permission errors: error per-target, continue others
- `own` mode sync with no source comments: warn that sync can't split back into partials, offer to write as single file

## Architecture

### New Files
- `src/manifest.go` — manifest parsing, validation, defaults merging
- `src/manifest_test.go` — manifest parsing tests
- `src/apply.go` — apply command logic (orchestrates build/own per target)
- `src/apply_test.go` — apply tests
- `src/sync.go` — sync command logic (reverse extraction)
- `src/sync_test.go` — sync tests
- `src/own.go` — "own" mode file writing (concatenate partials, write whole file)
- `src/own_test.go` — own mode tests
- `cmd/apply.go` — cobra subcommand for apply
- `cmd/remove_manifest.go` — cobra subcommand for manifest-driven remove
- `cmd/sync.go` — cobra subcommand for sync
- `cmd/init.go` — cobra subcommand for init

### Modified Files
- `go.mod` — add `gopkg.in/yaml.v3` dependency
- `cmd/root.go` — register new subcommands

### Reuse
- `merge` mode in apply directly reuses `PartialsBuildCommand`
- manifest-driven remove reuses `PartialsRemoveCommand` for merge targets
- Path expansion, comment style resolution, marker generation all reused as-is

## Implementation Priority

1. Manifest parsing + validation
2. `apply` command (merge mode only — reuses existing engine)
3. `apply` command (own mode)
4. Manifest-driven `remove`
5. `init` command
6. `sync` command (most complex, depends on source comments)
