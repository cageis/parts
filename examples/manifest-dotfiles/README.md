# Manifest Dotfiles Example

Demonstrates managing multiple dotfiles using a `.parts.yaml` manifest with both **merge** and **own** modes.

## Targets

| Name | Mode | Description |
|------|------|-------------|
| `ssh` | merge | SSH config — partials merged between markers, user content preserved |
| `bashrc` | own | Bashrc — entire file written from concatenated partials |
| `gitconfig` | own | Git config — entire file from partials |

## Running

```bash
# Build parts first
cd ../.. && make build && cd examples/manifest-dotfiles

# Run the example
./run-example.sh
```

## What It Demonstrates

1. **Apply all** — all 3 targets applied in one command
2. **Selective apply** — apply only specific targets
3. **Sync** — edit target file, sync changes back to partials
4. **Remove** — clean up: merge targets strip markers, own targets delete file
