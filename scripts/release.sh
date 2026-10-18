#!/bin/bash

# Release script for Docffy
# Usage: ./scripts/release.sh [major|minor|patch]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository"
    exit 1
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    print_warning "There are uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Get current version
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
print_status "Current version: $CURRENT_VERSION"

# Determine version bump type
BUMP_TYPE=${1:-patch}
if [[ ! "$BUMP_TYPE" =~ ^(major|minor|patch)$ ]]; then
    print_error "Invalid bump type. Use: major, minor, or patch"
    exit 1
fi

print_status "Bump type: $BUMP_TYPE"

# Calculate new version
IFS='.' read -r -a VERSION_PARTS <<< "${CURRENT_VERSION#v}"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
print_status "New version: $NEW_VERSION"

# Confirm release
read -p "Do you want to create release $NEW_VERSION? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Release cancelled"
    exit 0
fi

# Run tests
print_status "Running tests..."
if ! make test; then
    print_error "Tests failed. Aborting release."
    exit 1
fi
print_success "Tests passed"

# Build application
print_status "Building application..."
if ! make build; then
    print_error "Build failed. Aborting release."
    exit 1
fi
print_success "Build completed"

# Create git tag
print_status "Creating git tag..."
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
print_success "Git tag created: $NEW_VERSION"

# Push tag to remote
print_status "Pushing tag to remote..."
git push origin "$NEW_VERSION"
print_success "Tag pushed to remote"

# Build for all platforms
print_status "Building for all platforms..."
make build-all
print_success "Multi-platform builds completed"

# Create release notes
RELEASE_NOTES_FILE="RELEASE_NOTES_${NEW_VERSION#v}.md"
print_status "Creating release notes..."

cat > "$RELEASE_NOTES_FILE" << EOF
# Release $NEW_VERSION

## Changes

\`\`\`
$(git log --oneline "${CURRENT_VERSION}..HEAD")
\`\`\`

## Build Information

- Version: $NEW_VERSION
- Commit: $(git rev-parse HEAD)
- Build Time: $(date -u '+%Y-%m-%d %H:%M:%S UTC')

## Downloads

- Linux (AMD64): \`docffy-linux-amd64\`
- macOS (AMD64): \`docffy-darwin-amd64\`
- Windows (AMD64): \`docffy-windows-amd64.exe\`

## Installation

\`\`\`bash
# Download and run
chmod +x docffy-linux-amd64
./docffy-linux-amd64
\`\`\`

## Docker

\`\`\`bash
docker pull docffy:$NEW_VERSION
docker run -p 8080:8080 docffy:$NEW_VERSION
\`\`\`
EOF

print_success "Release notes created: $RELEASE_NOTES_FILE"

# Show version info
print_status "Version information:"
make version

print_success "Release $NEW_VERSION completed successfully!"
print_status "Next steps:"
echo "1. Review and update release notes: $RELEASE_NOTES_FILE"
echo "2. Create GitHub release with the generated notes"
echo "3. Upload the built binaries to the release"
echo "4. Update documentation if needed" 