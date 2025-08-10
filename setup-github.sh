#!/bin/bash

# Setup script for filesystem-exporter GitHub repository
# This script helps you set up the remote repository and push the code

set -e

echo "🚀 Setting up filesystem-exporter for GitHub..."

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "README.md" ]; then
    echo "❌ Error: This script must be run from the filesystem-exporter-new directory"
    exit 1
fi

# Check if git is initialized
if [ ! -d ".git" ]; then
    echo "❌ Error: Git repository not initialized"
    exit 1
fi

# Check if remote already exists
if git remote get-url origin >/dev/null 2>&1; then
    echo "⚠️  Remote 'origin' already exists:"
    git remote get-url origin
    read -p "Do you want to update it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git remote remove origin
    else
        echo "❌ Setup cancelled"
        exit 1
    fi
fi

# Add the remote
echo "📡 Adding remote repository..."
git remote add origin https://github.com/d0ugal/filesystem-exporter.git

echo "✅ Remote added successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Create the repository at https://github.com/d0ugal/filesystem-exporter"
echo "2. Push the code: git push -u origin main"
echo "3. Enable GitHub Actions in the repository settings"
echo "4. Set up branch protection rules for 'main'"
echo "5. Configure Release Please (optional)"
echo ""
echo "🔗 Repository URL: https://github.com/d0ugal/filesystem-exporter"
echo ""
echo "🎉 Ready to push! Run: git push -u origin main"
