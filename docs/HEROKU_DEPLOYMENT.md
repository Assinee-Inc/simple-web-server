# Heroku Deployment Guide

This guide covers the deployment process for Docffy to Heroku and common troubleshooting steps.

## Prerequisites

1. **Heroku CLI** installed: https://devcenter.heroku.com/articles/heroku-cli
2. **Git** repository initialized
3. **Go 1.23.4** or later installed

## Quick Deployment

Use the automated deployment script:

```bash
# Set your Heroku app name
export HEROKU_APP_NAME=your-app-name

# Run the deployment script
./scripts/deploy-heroku.sh
```

## Manual Deployment Steps

### 1. Create Heroku App

```bash
# Create a new Heroku app
heroku create your-app-name

# Or use an existing app
heroku git:remote -a your-app-name
```

### 2. Configure Build System

The app uses `heroku.yml` for deployment configuration. This file:
- Specifies Go as the build language
- Installs required packages (libpq-dev for PostgreSQL)
- Builds the application with proper flags
- Sets the correct environment variables

### 3. Set Environment Variables

```bash
# Required variables
heroku config:set APPLICATION_MODE=production
heroku config:set APPLICATION_NAME=Docffy
heroku config:set APP_KEY=your-secret-key-here

# Optional variables (set as needed)
heroku config:set MAIL_HOST=smtp.gmail.com
heroku config:set MAIL_USERNAME=your-email@gmail.com
heroku config:set MAIL_PASSWORD=your-app-password
heroku config:set S3_ACCESS_KEY=your-s3-access-key
heroku config:set S3_SECRET_KEY=your-s3-secret-key
heroku config:set STRIPE_SECRET_KEY=your-stripe-secret-key
heroku config:set HUB_DEVSENVOLVEDOR_TOKEN=your-token
```

### 4. Deploy

```bash
# Push to Heroku
git push heroku main

# Check logs
heroku logs --tail
```

## Configuration Files

### heroku.yml
- Defines the build process
- Specifies Go language and dependencies
- Sets build environment variables
- Defines the run command

### app.json
- App metadata for Heroku
- Environment variable definitions
- Add-on configurations (PostgreSQL)
- Formation settings

## Common Issues and Solutions

### 1. "No such file or directory" Error

**Problem**: `./bin/docffy: No such file or directory`

**Solution**: 
- Remove `Procfile` if it exists (use `heroku.yml` instead)
- Ensure `heroku.yml` is properly configured
- Check that the build process completes successfully

### 2. Database Connection Issues

**Problem**: Database connection fails

**Solution**:
- Ensure PostgreSQL add-on is provisioned: `heroku addons:create heroku-postgresql:mini`
- Check `DATABASE_URL` environment variable: `heroku config:get DATABASE_URL`
- Verify the app is in production mode: `heroku config:get APPLICATION_MODE`

### 3. Port Binding Issues

**Problem**: App fails to start due to port issues

**Solution**:
- Heroku automatically sets the `PORT` environment variable
- The app is configured to use this port automatically
- Check logs for port binding errors

### 4. Build Failures

**Problem**: Build process fails during deployment

**Solution**:
- Check Go version compatibility
- Ensure all dependencies are properly specified in `go.mod`
- Verify the build command in `heroku.yml`
- Check for missing files or incorrect paths

### 5. Environment Variable Issues

**Problem**: App fails due to missing environment variables

**Solution**:
- Set all required environment variables
- Use `heroku config` to view current settings
- Check the config loading logic in `internal/config/config.go`

## Monitoring and Debugging

### View Logs
```bash
# Real-time logs
heroku logs --tail

# Recent logs
heroku logs --num 100

# Logs for specific time
heroku logs --since 1h
```

### Run Commands
```bash
# Open bash shell
heroku run bash

# Run specific commands
heroku run go version
heroku run ls -la
```

### Check App Status
```bash
# App info
heroku info

# Dyno status
heroku ps

# Config variables
heroku config
```

## Performance Optimization

### 1. Database Optimization
- Use connection pooling
- Implement proper indexing
- Monitor query performance

### 2. Static Assets
- Ensure assets are properly served
- Consider CDN for static files
- Optimize asset sizes

### 3. Memory Usage
- Monitor memory consumption
- Optimize Go garbage collection
- Use appropriate dyno size

## Security Considerations

### 1. Environment Variables
- Never commit sensitive data
- Use Heroku config vars for secrets
- Rotate keys regularly

### 2. Database Security
- Use SSL connections
- Implement proper access controls
- Regular security updates

### 3. Application Security
- Enable HTTPS (automatic on Heroku)
- Implement rate limiting
- Use secure headers

## Troubleshooting Checklist

- [ ] Heroku CLI installed and logged in
- [ ] Git repository initialized and committed
- [ ] `heroku.yml` file present and correct
- [ ] No `Procfile` conflicting with `heroku.yml`
- [ ] All required environment variables set
- [ ] PostgreSQL add-on provisioned
- [ ] Build process completes successfully
- [ ] App starts without errors
- [ ] Database migrations run successfully
- [ ] Static assets accessible

## Support

If you encounter issues not covered in this guide:

1. Check Heroku logs: `heroku logs --tail`
2. Review the application logs for specific errors
3. Verify all configuration is correct
4. Test locally with production settings
5. Contact support with specific error messages and logs 