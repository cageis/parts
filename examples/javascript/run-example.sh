#!/bin/bash
set -e

echo "=== JavaScript Config Example ==="
echo "This example demonstrates merging JavaScript config partials using slash comments."
echo

# Copy original for safety
cp app.js app.js.backup

echo "Original app.js:"
cat app.js
echo

echo "Partials to merge:"
echo "- database.js:"
cat partials/database.js
echo
echo "- api.js:"
cat partials/api.js
echo

echo "Running: parts app.js partials \"//\""
../../bin/parts app.js partials "//"

echo "Result:"
cat app.js
echo

echo "=== Auto-detection example ==="
echo "Running: parts --dry-run app.js partials \"auto\""
../../bin/parts --dry-run app.js partials "auto"

# Restore original
mv app.js.backup app.js
echo
echo "✅ JavaScript config example completed!"