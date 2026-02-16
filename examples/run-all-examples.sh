#!/bin/bash
set -e

echo "🚀 Running all Parts examples..."
echo "================================="

# Build parts first
echo "Building parts..."
cd ..
make build
cd examples

echo
echo "📁 SSH Config Example"
echo "====================="
cd ssh-config
./run-example.sh
cd ..

echo
echo "📁 Hosts File Example"
echo "====================="
cd hosts
./run-example.sh
cd ..

echo
echo "📁 JavaScript Config Example"
echo "============================"
cd javascript
./run-example.sh
cd ..

echo
echo "📁 CSS Styles Example"
echo "===================="
cd css
./run-example.sh
cd ..

echo
echo "📁 Python Config Example"
echo "========================"
cd python
./run-example.sh
cd ..

echo
echo "📁 SQL Schema Example"
echo "===================="
cd sql
./run-example.sh
cd ..

echo
echo "📁 Manifest Dotfiles Example"
echo "============================"
cd manifest-dotfiles
./run-example.sh
cd ..

echo
echo "📁 Manifest Sync Workflow Example"
echo "================================="
cd manifest-sync-workflow
./run-example.sh
cd ..

echo
echo "📁 Manifest Migration Example"
echo "=============================="
cd manifest-migration
./run-example.sh
cd ..

echo
echo "🎉 All examples completed successfully!"
echo "======================================="