# Implementation Plan: Fix Creator Account Confirmation Mismatch

**Branch**: `008-fix-creator-account-confirmation` | **Date**: 2026-06-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-fix-creator-account-confirmation/spec.md`

## Summary

The email confirmation flow redirects verified users to `/stripe-connect/welcome` where a lookup for their associated creator profile is performed. This lookup currently uses a JOIN query linking `users.id` with `creators.user_id`. When the database has a creator record for the email but with a mismatched `user_id` (e.g. from local environment DB truncations/seeds), the lookup fails with HTTP 404/record not found, locking the user out of onboarding.

The proposed solution updates the creator lookup to auto-heal this state: if the JOIN fails but the creator exists directly under the same email, the system automatically links it to the current user's ID, logs the occurrence, and returns the healed profile.

## Technical Context

**Language/Version**: Go 1.23.4  
**Primary Dependencies**: GORM, Chi Router  
**Storage**: PostgreSQL (production & local Docker)  
**Testing**: go test, testify  
**Target Platform**: Linux server (production) / macOS (local development)  
**Project Type**: web-service  
**Performance Goals**: Instant lookup (< 50ms)  
**Constraints**: Keep creator and user data strictly aligned by email  
**Scale/Scope**: All user email verification & onboarding redirects

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Core Principles**:
  - **III. Data Validation**: Data integrity is preserved by resolving mismatched associations automatically when email identities are proven through confirmation.
  - **IV. Testability**: The auto-healing logic will be fully tested with a new repository unit test suite.
  - **V. Modularity**: The fix will reside in the repository layer where the query is performed, adhering to the hander -> service -> repository modularity pattern.

## Project Structure

### Documentation (this feature)

```text
specs/008-fix-creator-account-confirmation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code

```text
internal/
├── account/
│   ├── model/
│   │   └── creator.go
│   ├── repository/
│   │   ├── creator_gorm.go
│   │   └── creator_gorm_test.go      # [NEW]
│   └── service/
│       └── creator_service.go
├── auth/
│   └── service/
│       └── user_service.go
pkg/
└── database/
    └── database.go
```

**Structure Decision**: Single-project Go package structure. Real directories modified are `internal/account/repository/` and `pkg/database/`.

## Proposed Changes

### Database Initialization

#### [MODIFY] [database.go](file:///Users/ang/Documents/simple-web-server/pkg/database/database.go)
- Ensure `config.LoadConfigs()` is called inside `Connect()` so that tests running locally outside Docker can successfully parse `DATABASE_URL` and initialize the PostgreSQL database.

### Account Repository

#### [MODIFY] [creator_gorm.go](file:///Users/ang/Documents/simple-web-server/internal/account/repository/creator_gorm.go)
- Refactor `FindCreatorByUserEmail()` to:
  - Replace `database.DB` direct usage with `cr.db` to support mocking/alternate test databases.
  - On query failure due to `gorm.ErrRecordNotFound`, query `creators` by `email` directly, and `users` by `email` directly.
  - If both exist and `creator.user_id != user.id`, log the healing action, update the creator's `user_id` to match the user's correct `id`, save the record, and return it.

#### [NEW] [creator_gorm_test.go](file:///Users/ang/Documents/simple-web-server/internal/account/repository/creator_gorm_test.go)
- Add a new repository unit test suite using `testify/suite`.
- Implement test cases:
  1. `TestFindCreatorByUserEmail_Success_Aligned`: Aligned user and creator.
  2. `TestFindCreatorByUserEmail_Success_AutoHealMismatched`: Mismatched `user_id` is auto-healed and successfully returned.
  3. `TestFindCreatorByUserEmail_NotFound`: Returns error when creator doesn't exist.

## Verification Plan

### Automated Tests
Run repository and service tests to verify the fix works:
```bash
DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=simple_web_server sslmode=disable" go test ./internal/account/... ./internal/auth/...
```

### Manual Verification
1. Insert a user record with email `test-heal@example.com` and ID `55`.
2. Insert a creator record with email `test-heal@example.com` and `user_id = 99` (mismatched).
3. Access `/stripe-connect/welcome` with a session matching the user's email.
4. Verify the page loads successfully and the database record in `creators` is updated to `user_id = 55`.
