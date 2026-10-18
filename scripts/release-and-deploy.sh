#!/bin/bash

# Release and Deploy script for Docffy
# Usage: ./scripts/release-and-deploy.sh [major|minor|patch] [heroku-app-name]

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

# Get parameters
BUMP_TYPE=${1:-patch}
HEROKU_APP=${2:-}

if [[ -z "$HEROKU_APP" ]]; then
    print_error "Please provide Heroku app name: ./scripts/release-and-deploy.sh [major|minor|patch] [heroku-app-name]"
    exit 1
fi

print_status "Starting release and deploy process..."
print_status "Bump type: $BUMP_TYPE"
print_status "Heroku app: $HEROKU_APP"

# Step 1: Run the release script
print_status "Step 1: Creating release..."
if ! ./scripts/release.sh "$BUMP_TYPE"; then
    print_error "Release failed. Aborting deploy."
    exit 1
fi

# Get the new version from the release script
NEW_VERSION=$(git describe --tags --abbrev=0)
print_success "Release $NEW_VERSION created successfully"

# Step 2: Push to origin (including tags)
print_status "Step 2: Pushing to origin..."
git push origin master
git push origin "$NEW_VERSION"

# Step 3: Deploy to Heroku (Heroku will build the binary)
print_status "Step 3: Deploying to Heroku..."
if ! git push heroku master; then
    print_error "Heroku deploy failed."
    exit 1
fi

# Step 4: Verify deployment
print_status "Step 4: Verifying deployment..."
sleep 5  # Wait for deployment to complete

# Check if app is running
if heroku ps --app "$HEROKU_APP" | grep -q "web.1.*up"; then
    print_success "Deployment verified successfully"
else
    print_warning "Deployment verification failed. Check logs with: heroku logs --tail --app $HEROKU_APP"
fi

# Step 5: Show deployment info
print_status "Step 5: Deployment information:"
echo "Version: $NEW_VERSION"
echo "App URL: https://$HEROKU_APP.herokuapp.com"
echo "Logs: heroku logs --tail --app $HEROKU_APP"
echo "Status: heroku ps --app $HEROKU_APP"

print_success "Release and deploy completed successfully!"
print_status "Release $NEW_VERSION is now live on Heroku!" 