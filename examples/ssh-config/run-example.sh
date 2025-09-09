#!/bin/bash
set -e

echo "=== SSH Config Example ==="
echo "This example demonstrates merging SSH config partials using hash comments."
echo

# Copy original for safety
cp config config.backup

echo "Original config:"
cat config
echo

echo "Partials to merge:"
echo "- work-servers:"
cat partials/work-servers
echo
echo "- development:"
cat partials/development
echo

echo "Running: parts config partials \"#\""
../../bin/parts config partials "#"

echo "Result:"
cat config
echo

echo "=== Dry run example ==="
echo "Running: parts --dry-run config partials \"#\""
../../bin/parts --dry-run config partials "#"

# Restore original
mv config.backup config
echo
echo "✅ SSH config example completed!"