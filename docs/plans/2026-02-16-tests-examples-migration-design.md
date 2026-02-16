# Tests, Examples & Migration — Design

**Goal:** Fill test coverage gaps for the manifest workflow, add runnable examples demonstrating new features, and provide a migration path from legacy CLI to manifest mode via `init --from`.

---

## 1. `init --from` Migration Feature

### Usage

```bash
# Create new manifest with a target
parts init --from ~/.ssh/config ~/.ssh/config.d "#" --name ssh

# Append another target to existing manifest
parts init --from ~/.vimrc ./vim/ "auto" --name vimrc --mode own
```

### Behavior

- **No `.parts.yaml` exists:** Creates one with the specified target (no skeleton comments — real content).
- **`.parts.yaml` exists:** Reads it, appends the new target, writes it back.
- `--name` is optional. If omitted, derive from target filename: `~/.ssh/config` → `ssh-config`, `~/.vimrc` → `vimrc`.
- `--mode` defaults to `merge`.
- Validates that the partials directory exists.
- Errors if the target name already exists in the manifest.

### Files

- Modify: `cmd/init.go`
- Create: `cmd/init_from_test.go`

---

## 2. New Tests

### `cmd/init_from_test.go` — Migration tests

| Test | Description |
|------|-------------|
| `TestInitFrom_CreatesNewManifest` | No existing manifest → creates one with the target |
| `TestInitFrom_AppendsToExisting` | Existing manifest → adds new target |
| `TestInitFrom_DeriveNameFromPath` | Omit `--name` → auto-derived name |
| `TestInitFrom_ExplicitName` | `--name custom` → name used |
| `TestInitFrom_ValidatesPartialsDir` | Missing partials dir → error |
| `TestInitFrom_DuplicateTargetName` | Name collision → error |
| `TestInitFrom_GeneratedManifestIsUsable` | init --from → apply → works |

### `cmd/workflow_test.go` — End-to-end integration

| Test | Description |
|------|-------------|
| `TestWorkflow_InitFromApplyRemoveRoundTrip` | init --from → apply → remove → clean state |
| `TestWorkflow_ApplyEditSyncReapply` | apply → edit target → sync → re-apply → round-trip |
| `TestWorkflow_MixedModesApplyAndRemove` | 3+ targets, merge + own, apply all, remove all |
| `TestWorkflow_SelectiveOperations` | selective apply/sync/remove, others untouched |

### `src/sync_test.go` — Sync edge cases (additions)

| Test | Description |
|------|-------------|
| `TestExtractPartialSections_BlockComments` | `/* Source: path */` parsing |
| `TestExtractPartialSections_HTMLComments` | `<!-- Source: path -->` parsing |
| `TestSyncTarget_OwnMode` | Sync in own mode |
| `TestSyncTarget_MissingTargetFile` | Error when target doesn't exist |
| `TestSyncTarget_NoManagedSection` | Merge mode with no markers → 0 updates |

### `src/own_test.go` — Own mode edge cases (additions)

| Test | Description |
|------|-------------|
| `TestPartialsOwnCommand_EmptyPartialsDir` | Zero partials → empty file |
| `TestPartialsOwnCommand_MultiplePartialsOrdering` | 5+ partials → alphabetical |
| `TestPartialsOwnCommand_NoTrailingNewline` | Newline normalization |

### `cmd/init_test.go` — Init validation (addition)

| Test | Description |
|------|-------------|
| `TestInitCommand_GeneratedManifestIsParseable` | Skeleton is valid YAML |

### `src/manifest_test.go` — Manifest edge cases (additions)

| Test | Description |
|------|-------------|
| `TestLoadManifest_TildePaths` | `~/` paths parse correctly |
| `TestLoadManifest_EmptyTargetMap` | `targets:` with no children → error |
| `TestLoadManifest_SpecialTargetNames` | Hyphens, underscores, dots work |

---

## 3. New Examples

### `examples/manifest-dotfiles/`

Multi-target manifest managing a dotfile repository.

```
examples/manifest-dotfiles/
  ├── .parts.yaml          # 3 targets: ssh (merge), bashrc (own), gitconfig (own)
  ├── ssh/
  │   ├── work             # Host work config
  │   └── personal         # Host personal config
  ├── bash/
  │   ├── 01-exports       # Environment variables
  │   └── 02-aliases       # Shell aliases
  ├── git/
  │   └── gitconfig        # Git configuration
  ├── run-example.sh       # Demonstrates: apply all → selective → edit → sync → remove
  └── README.md
```

### `examples/manifest-sync-workflow/`

Bidirectional sync demonstration.

```
examples/manifest-sync-workflow/
  ├── .parts.yaml          # 1 merge target: ssh config
  ├── ssh/
  │   ├── work             # Work SSH config
  │   └── staging          # Staging SSH config
  ├── run-example.sh       # Demonstrates: apply → edit target → sync → re-apply round-trip
  └── README.md
```

### `examples/manifest-migration/`

Migration from legacy CLI to manifest mode.

```
examples/manifest-migration/
  ├── partials/
  │   └── work.conf        # Example partial
  ├── run-example.sh       # Demonstrates: legacy usage → init --from → manifest apply
  └── README.md
```

---

## Out of Scope

- Performance/benchmark tests (v2.0)
- Concurrency/race condition tests
- Backup functionality (not yet implemented)
- `--detect` auto-scan mode for init
