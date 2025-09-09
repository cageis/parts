#!/bin/bash
set -e

echo "=== Python Config Example ==="
echo "This example demonstrates auto-detection of comment style based on file extension."
echo

# Copy original for safety
cp config.py config.py.backup

echo "Original config.py:"
cat config.py
echo

echo "Partials to merge:"
echo "- logging.py:"
cat partials/logging.py
echo
echo "- cache.py:"
cat partials/cache.py
echo

echo "Running: parts config.py partials \"auto\""
echo "(Auto-detection will use # comments for .py files)"
../../bin/parts config.py partials "auto"

echo "Result:"
cat config.py
echo

# Restore original
mv config.py.backup config.py
echo "✅ Python config example completed!"