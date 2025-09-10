#!/bin/bash

echo "======================================"
echo "Parts - Hosts File Example"
echo "======================================"
echo

# Build the parts binary if it doesn't exist
if [ ! -f "../../bin/parts" ]; then
    echo "Building parts binary..."
    cd ../.. && make build && cd examples/hosts
    echo
fi

echo "Original hosts file:"
echo "--------------------------------------"
cat hosts
echo
echo "======================================"

echo "Partial files in partials/ directory:"
echo "--------------------------------------"
for partial in partials/*; do
    echo "=== $(basename "$partial") ==="
    cat "$partial"
    echo
done

echo "======================================"
echo "Running: ../../bin/parts hosts ./partials \"#\""
echo "======================================"
echo

# Run parts to merge the partials
../../bin/parts hosts ./partials "#"

echo "Result after merging partials:"
echo "--------------------------------------"
cat hosts
echo

echo "======================================"
echo "Cleaning up - removing partials section"
echo "======================================"
echo

# Clean up by removing the partials section
../../bin/parts --remove hosts "#"

echo "Final result (back to original):"
echo "--------------------------------------"
cat hosts
echo

echo "======================================"
echo "Hosts file example complete!"
echo "======================================"