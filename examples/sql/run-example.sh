#!/bin/bash
set -e

echo "=== SQL Schema Example ==="
echo "This example demonstrates merging SQL partials using dash comments."
echo

# Copy original for safety
cp schema.sql schema.sql.backup

echo "Original schema.sql:"
cat schema.sql
echo

echo "Partials to merge:"
echo "- products.sql:"
cat partials/products.sql
echo
echo "- orders.sql:"
cat partials/orders.sql
echo

echo "Running: parts schema.sql partials \"auto\""
echo "(Auto-detection will use -- comments for .sql files)"
../../bin/parts schema.sql partials "auto"

echo "Result:"
cat schema.sql
echo

# Restore original
mv schema.sql.backup schema.sql
echo "✅ SQL schema example completed!"