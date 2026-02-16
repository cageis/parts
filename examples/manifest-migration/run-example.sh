#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARTS="$SCRIPT_DIR/../../bin/parts"
cd "$SCRIPT_DIR"

# Clean previous runs
rm -f .parts.yaml
rm -rf output

echo "=== Migration Example ==="
echo "Demonstrates migrating from legacy CLI to manifest mode."
echo

mkdir -p output
echo "# SSH Config" > output/ssh-config
echo "" >> output/ssh-config

echo "--- Step 1: Legacy CLI usage ---"
echo "Running: parts output/ssh-config partials \"#\""
$PARTS output/ssh-config partials "#"
echo
echo "Result (legacy):"
cat output/ssh-config
echo

echo "--- Step 2: Remove legacy partials ---"
$PARTS --remove output/ssh-config "#"
echo "Cleaned target file."
echo

echo "--- Step 3: Migrate to manifest with init --from ---"
echo "Running: parts init --from output/ssh-config --from ./partials --from \"#\" --name ssh"
$PARTS init --from output/ssh-config --from ./partials --from "#" --name ssh
echo
echo "Generated .parts.yaml:"
cat .parts.yaml
echo

echo "--- Step 4: Apply using manifest ---"
$PARTS apply
echo
echo "Result (manifest):"
cat output/ssh-config
echo

echo "--- Step 5: Cleanup ---"
$PARTS remove
rm -f .parts.yaml
rm -rf output

echo
echo "=== Migration example completed ==="
