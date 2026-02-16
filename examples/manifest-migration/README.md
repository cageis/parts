# Migration Example

Demonstrates migrating from the legacy positional-argument CLI to the new manifest-driven workflow.

## Legacy vs Manifest

**Legacy (positional args):**
```bash
parts ~/.ssh/config ~/.ssh/config.d "#"
```

**Manifest (`.parts.yaml`):**
```bash
parts init --from ~/.ssh/config --from ./ssh --from "#" --name ssh
parts apply
```

## Running

```bash
cd ../.. && make build && cd examples/manifest-migration
./run-example.sh
```

## Migration Steps

1. Run your existing legacy command one last time
2. Remove the partials section: `parts --remove <file> <comment>`
3. Create manifest: `parts init --from <file> --from <dir> --from <comment> --name <name>`
4. Apply with manifest: `parts apply`
