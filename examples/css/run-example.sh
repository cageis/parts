#!/bin/bash
set -e

echo "=== CSS Styles Example ==="
echo "This example demonstrates merging CSS partials using block comments."
echo

# Copy original for safety
cp styles.css styles.css.backup

echo "Original styles.css:"
cat styles.css
echo

echo "Partials to merge:"
echo "- buttons.css:"
cat partials/buttons.css
echo
echo "- forms.css:"
cat partials/forms.css
echo

echo "Running: parts styles.css partials \"/*\""
../../bin/parts styles.css partials "/*"

echo "Result:"
cat styles.css
echo

# Restore original
mv styles.css.backup styles.css
echo "✅ CSS styles example completed!"