# Manifest Sync Workflow Example

Demonstrates bidirectional sync between a target file and its partial source files.

## Workflow

1. **Apply** — merge partials into the target
2. **Edit** — make changes directly in the target file
3. **Sync** — pull changes back into the individual partial files
4. **Re-apply** — verify the round-trip preserves changes

## Running

```bash
cd ../.. && make build && cd examples/manifest-sync-workflow
./run-example.sh
```

## Key Concept

The sync feature uses `# Source: <path>` comments to map sections of the target file back to their source partial files. When you edit the target and run `parts sync`, the changes flow back to the correct partial.
