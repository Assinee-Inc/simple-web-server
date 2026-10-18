# Tasks: Fix Creator Account Confirmation Mismatch

**Input**: Design documents from `/specs/008-fix-creator-account-confirmation/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verification of current environment

- [x] T001 Verify all existing tests pass locally by running `go test ./...` with the SQLite database fallback.

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core configuration alignment for tests

- [x] T002 Update database connector in `pkg/database/database.go` to invoke `config.LoadConfigs()` inside `Connect()` to enable proper configuration loading in repository test suites.

## Phase 3: User Story 1 - Auto-healing Creator Linkage (Priority: P1) 🎯 MVP

**Goal**: Automatically heal user-creator association during lookup if emails match but user IDs are misaligned.

**Independent Test**: Manually align a creator's email with a user's email, but set a mismatched `user_id` in the `creators` table. Querying the creator by user email must update the `user_id` and return the creator record successfully.

### Tests for User Story 1

- [x] T003 [P] [US1] Create unit tests in `internal/account/repository/creator_gorm_test.go` covering successful query, auto-healing of mismatched user ID, and not-found handling.

### Implementation for User Story 1

- [x] T004 [US1] Implement the auto-healing fallback logic in `FindCreatorByUserEmail` in `internal/account/repository/creator_gorm.go`.

## Phase 4: User Story 2 - Standard Account Confirmation (Priority: P2)

**Goal**: Confirm that standard aligned user-creator confirmation continues to function correctly.

**Independent Test**: Register a new creator via browser or client requests, verify that the email confirmation token works, and the user is redirected to the onboarding screen.

### Implementation for User Story 2

- [x] T005 [US2] Verify alignment implementation in `RegisterCreatorSSR` in `internal/account/handler/creator_handler.go` and onboarding redirect in `OnboardingWelcome` in `internal/account/handler/stripe_connect_handler.go`.

## Phase 5: User Story 3 - Service and Handler Verification (Priority: P3)

**Goal**: Validate the entire registration and confirmation flow services and endpoints.

**Independent Test**: Run service-level tests for users and handlers.

### Tests for User Story 3

- [x] T006 [P] [US3] Verify that the tests in `internal/auth/service/user_service_test.go` and `internal/auth/handler/auth_handler_test.go` continue to pass.

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final project cleanup and verification

- [x] T007 Run the full Go test suite locally using PostgreSQL: `DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=simple_web_server sslmode=disable" go test ./...`
- [x] T008 [P] Document implemented changes in `walkthrough.md`.

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1 completion.
- **User Story 1 (Phase 3)**: Depends on Phase 2 completion.
- **User Story 2 (Phase 4)**: Depends on Phase 3 completion.
- **User Story 3 (Phase 5)**: Depends on Phase 3 completion.
- **Polish (Phase 6)**: Depends on all stories being complete.

### Parallel Opportunities

- T003 can be implemented in parallel with T004 as it is a test file.
- T006 can be run in parallel with other tests.
