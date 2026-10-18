#!/bin/bash

# Heroku Deployment Script for Docffy
# This script helps ensure proper deployment to Heroku

set -e

echo "🚀 Starting Heroku deployment process..."

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from the project root."
    exit 1
fi

# Check if heroku CLI is installed
if ! command -v heroku &> /dev/null; then
    echo "❌ Error: Heroku CLI not found. Please install it first:"
    echo "   https://devcenter.heroku.com/articles/heroku-cli"
    exit 1
fi

# Check if we have a git repository
if [ ! -d ".git" ]; then
    echo "❌ Error: Not a git repository. Please initialize git first."
    exit 1
fi

# Ensure we're using heroku.yml for deployment
if [ -f "Procfile" ]; then
    echo "⚠️  Warning: Procfile found. Removing it to use heroku.yml instead."
    rm Procfile
fi

# Check if heroku.yml exists
if [ ! -f "heroku.yml" ]; then
    echo "❌ Error: heroku.yml not found. Please create it first."
    exit 1
fi

# Build the application locally to test
echo "🔨 Building application locally..."
make build-prod

# Check if build was successful
if [ ! -f "bin/docffy-linux-amd64" ]; then
    echo "❌ Error: Build failed. Please check the build process."
    exit 1
fi

echo "✅ Build successful!"

# Check if we have a Heroku app configured
if [ -z "$HEROKU_APP_NAME" ]; then
    echo "⚠️  Warning: HEROKU_APP_NAME not set. You'll need to specify the app name manually."
    echo "   Usage: HEROKU_APP_NAME=your-app-name ./scripts/deploy-heroku.sh"
    echo ""
    echo "Available Heroku apps:"
    heroku apps
    echo ""
    echo "To set up a new app:"
    echo "   heroku create your-app-name"
    echo "   heroku git:remote -a your-app-name"
    exit 1
fi

# Deploy to Heroku
echo "🚀 Deploying to Heroku app: $HEROKU_APP_NAME"

# Ensure we're connected to the right Heroku app
heroku git:remote -a $HEROKU_APP_NAME

# Push to Heroku
echo "📤 Pushing to Heroku..."
git push heroku main

# Check the deployment status
echo "🔍 Checking deployment status..."
heroku logs --tail --app $HEROKU_APP_NAME

echo "✅ Deployment process completed!"
echo ""
echo "📋 Next steps:"
echo "1. Check the logs above for any errors"
echo "2. Open your app: https://$HEROKU_APP_NAME.herokuapp.com"
echo "3. Set up environment variables if needed:"
echo "   heroku config:set APPLICATION_MODE=production --app $HEROKU_APP_NAME"
echo "   heroku config:set APPLICATION_NAME=Docffy --app $HEROKU_APP_NAME"
echo "   heroku config:set APP_KEY=your-secret-key --app $HEROKU_APP_NAME"
echo ""
echo "🔧 Useful commands:"
echo "   heroku logs --tail --app $HEROKU_APP_NAME  # View live logs"
echo "   heroku run bash --app $HEROKU_APP_NAME     # Run bash in Heroku"
echo "   heroku config --app $HEROKU_APP_NAME       # View config vars" 