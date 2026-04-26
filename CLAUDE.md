# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run          # Start dev server (go run cmd/web/main.go) on :8080
make dev          # Hot reload via air
make test         # go test ./...
make build        # Binary to bin/simple-web-server
make build-prod   # Linux AMD64 production binary
make css          # Compile Tailwind CSS
make css-watch    # Watch Tailwind CSS
make setup-env    # Interactive .env setup from template
make deps         # go mod download + go mod tidy
```

**Single test:** `go test ./internal/sales/handler/... -v -run TestMyFunc`

**Integration tests** use SQLite in-memory via `gorm.io/driver/sqlite` — look for `*_integration_test.go`.

## Architecture

**Entry point:** `cmd/web/main.go` — manual constructor injection (no DI framework). All repositories, services, and handlers are wired here.

**Module structure** under `internal/`:
- `account/` — Creator profiles, Stripe Connect onboarding, dashboard
- `auth/` — Login, signup, session, email verification, password reset
- `delivery/` — Download logs and file delivery
- `library/` — Ebook and file management, watermarking (pdfcpu/unipdf)
- `sales/` — Purchases, transactions, clients, Stripe checkout + webhook
- `subscription/` — Stripe subscriptions
- `shared/handler/` — Error and home handlers; `shared/web/` — flash messages, utils

Each module follows: `model/ → repository/ → service/ → handler/`. Interfaces are defined in the service/repository layer; GORM implementations live under `repository/gorm/`.

**Mocks** are centralized in `internal/mocks/` (not per-module). Use `MockSalesEmailService` for sales email, `MockAuthEmailService` for auth email.

**Templates** live in `web/pages/{module}/{page}.html` and `web/layouts/`. The renderer is `pkg/template` which automatically injects the authenticated user, flash messages (from cookies), form data, CSRF token, and subscription status into every template. Call with `h.templateRenderer.View(w, r, "module/page", data, "layout-name")`.

**Active layouts:**
- `guest.html` — DaisyUI 4 (auth pages)
- `admin-daisy.html` — DaisyUI 4 (private pages, drawer sidebar)
- `landing.html` — DaisyUI 4 (home)
- `admin.html` — Bootstrap (legacy, pending migration to DaisyUI)

## Database

GORM with SQLite in development (`./mydb.db`) and PostgreSQL in production (detected via `APPLICATION_MODE=production`). `AutoMigrate` runs on startup in `pkg/database/database.go`. All models carry a semantic `PublicID` with a prefix (e.g., `pur_`, `cli_`, `ebk_`) — never expose the raw auto-increment ID in URLs.

## Stripe Integration

Purchases flow: `CheckoutHandler.CreateCheckoutSession` → Stripe redirect → `PurchaseSuccessView` (UX only) + Stripe webhook `checkout.session.completed` → `StripeHandler.handleEbookPayment`.

**The webhook is the single authoritative trigger** for post-payment side-effects (email, transaction recording). `PurchaseSuccessView` must not duplicate those actions.

## Session & Auth

Gorilla sessions (AES-256 + HMAC-SHA-256). Keys from `SESSION_AUTH_KEY` / `SESSION_ENC_KEY` env vars. Private routes go through `authmw.AuthMiddleware` (validates session, CSRF, email verification) → `StripeOnboardingMiddleware` → `SubscriptionMiddleware`.

## Key Config

`internal/config/config.go` loads from `.env` in dev, env vars in production. Business fee constants (platform fee %, Stripe processing fee) are in `internal/config/business.go`. Feature flags: `HIDE_EBOOK_AUTHOR_FIELD`, `HIDE_RESEND_LINK`.

## Important pkg/ Packages

- `pkg/middleware/` — `SecurityHeaders` (CSP, X-Frame-Options, etc.), `RateLimiter` (IP-based, in-memory)
- `pkg/gov/` — CPF validation via Receita Federal / Hub Desenvolvedor API
- `pkg/cookie/` — Flash message encoding/decoding
- `pkg/storage/` — AWS S3 uploads
- `pkg/utils/` — bcrypt encrypter, UUID/token generators, money formatting

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
