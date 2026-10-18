# Quickstart: Fix Creator Account Confirmation

This document explains how to set up, test, and verify the creator auto-healing confirmation flow.

## Prerequisites

Ensure you have:
1. Go 1.23.4 installed.
2. PostgreSQL running in Docker (via `make up` or `docker compose up -d`).

## Environment Configuration

When running tests locally outside of the Docker container, you must configure `DATABASE_URL` to point to `localhost` instead of `postgres`.

Run the test suite with the following environment variables:

```bash
DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=simple_web_server sslmode=disable" go test ./...
```

## Running Verification Tests

To verify the auto-healing and repository fixes, run the specific test files:

```bash
# Run account repository and service tests
DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=simple_web_server sslmode=disable" go test ./internal/account/... ./internal/auth/...
```
