#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARTS="$SCRIPT_DIR/../../bin/parts"
cd "$SCRIPT_DIR"

echo "=== Sync Workflow Example ==="
echo "Demonstrates bidirectional sync between target and partials."
echo

mkdir -p output
echo "# SSH Config - managed by parts" > output/ssh-config
echo "" >> output/ssh-config

echo "--- Step 1: Initial apply ---"
$PARTS apply
echo
echo "Target file:"
cat output/ssh-config
echo

echo "--- Step 2: Simulate editing the target file ---"
# Change port for hosts (macOS-compatible sed)
if sed -i '' 's/Port 22/Port 2222/g' output/ssh-config 2>/dev/null; then
    true
else
    sed -i 's/Port 22/Port 2222/g' output/ssh-config
fi
echo "Changed all Port 22 -> Port 2222 in target"
echo
echo "Modified target:"
cat output/ssh-config
echo

echo "--- Step 3: Sync changes back to partials ---"
$PARTS sync
echo
echo "Work partial after sync:"
cat ssh/work
echo
echo "Staging partial after sync:"
cat ssh/staging
echo

echo "--- Step 4: Re-apply to verify round-trip ---"
$PARTS apply
echo
echo "Target after re-apply (should match step 2 edit):"
cat output/ssh-config
echo

echo "--- Step 5: Cleanup ---"
$PARTS remove
rm -rf output

# Restore original partials
cat > ssh/work << 'PARTIAL'
Host work
    HostName work.example.com
    User deploy
    Port 22
PARTIAL

cat > ssh/staging << 'PARTIAL'
Host staging
    HostName staging.example.com
    User deploy
    Port 22
PARTIAL

echo
echo "=== Sync workflow example completed ==="
