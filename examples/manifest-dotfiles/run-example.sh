#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARTS="$SCRIPT_DIR/../../bin/parts"
cd "$SCRIPT_DIR"

echo "=== Manifest Dotfiles Example ==="
echo "Demonstrates multi-target manifest with merge + own modes."
echo

# Create base SSH config (merge mode preserves this)
mkdir -p output
echo "# My SSH Config" > output/ssh-config
echo "# Custom settings above will be preserved" >> output/ssh-config
echo "" >> output/ssh-config

echo "--- Step 1: Apply all targets ---"
$PARTS apply
echo

echo "SSH config (merge mode - original content preserved):"
cat output/ssh-config
echo

echo "Bashrc (own mode - entirely from partials):"
cat output/bashrc
echo

echo "Gitconfig (own mode):"
cat output/gitconfig
echo

echo "--- Step 2: Selective apply (ssh only) ---"
# First remove to reset
$PARTS remove
echo

# Re-create base SSH config
echo "# My SSH Config" > output/ssh-config
echo "# Custom settings above will be preserved" >> output/ssh-config
echo "" >> output/ssh-config

$PARTS apply ssh
echo

echo "SSH config updated:"
cat output/ssh-config
echo

echo "Bashrc should not exist yet:"
ls output/bashrc 2>/dev/null && echo "(exists)" || echo "(not created - correct)"
echo

echo "--- Step 3: Apply remaining targets ---"
$PARTS apply bashrc gitconfig
echo

echo "--- Step 4: Sync demo (edit target, sync back) ---"
# Simulate user editing the SSH config (macOS-compatible sed)
if sed -i '' 's/Port 22/Port 2222/g' output/ssh-config 2>/dev/null; then
    true
else
    sed -i 's/Port 22/Port 2222/g' output/ssh-config
fi
echo "Edited SSH config (changed Port 22 -> 2222)"
$PARTS sync ssh
echo "Partial updated:"
cat ssh/work
echo

echo "--- Step 5: Remove all ---"
$PARTS remove
echo

echo "SSH config after remove (original content preserved):"
cat output/ssh-config
echo

echo "Bashrc after remove (deleted):"
ls output/bashrc 2>/dev/null && echo "(still exists - error)" || echo "(deleted - correct)"
echo

# Cleanup
rm -rf output

# Restore original partials
cat > ssh/work << 'PARTIAL'
Host work
    HostName work.example.com
    User deploy
    Port 22
    IdentityFile ~/.ssh/id_work
PARTIAL

echo "=== Example completed ==="
