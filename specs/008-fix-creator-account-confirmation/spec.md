# Feature Specification: Fix Creator Account Confirmation

**Feature Branch**: `008-fix-creator-account-confirmation`  
**Created**: 2026-06-13  
**Status**: Draft  
**Input**: User description: "O Fluxo de confirmação de conta está falhando por não encontrar os dados do Creator vinculado ao email do usuário. Adicione testes para evitar que este cenário aconteça novamente."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Auto-healing Creator Linkage during Account Confirmation (Priority: P1)

As a registered creator, when I click on the account confirmation link, I want the system to successfully activate my account and redirect me to the welcome page, even if the database has a creator record with my email but a mismatched user ID (e.g., due to local database resets, seed shifts, or partial imports).

**Why this priority**: Mismatched user and creator records completely lock users out of their accounts during email confirmation, rendering the system unusable for them. Healing this state automatically provides a seamless user experience.

**Independent Test**: Can be tested by manually inserting a user with email `user@example.com` and a creator record with the same email but a mismatched `user_id` (e.g., `user_id = 999`). Triggering the email confirmation endpoint for that user must successfully verify the user's email, automatically update the creator's `user_id` to match the user's correct database ID, and redirect the user to `/stripe-connect/welcome` without errors.

**Acceptance Scenarios**:

1. **Given** a user with email `user@example.com` has a verification token, and a creator record with email `user@example.com` exists but points to a mismatched `user_id`, **When** the user hits the account confirmation endpoint with the token, **Then** their email is marked as verified, **And** the creator record's `user_id` is updated to the user's actual ID, **And** the user is redirected to `/stripe-connect/welcome` successfully.
2. **Given** the user is redirected to `/stripe-connect/welcome`, **When** they load the welcome page, **Then** they see their creator profile information instead of a "Creator não encontrado" error.

---

### User Story 2 - Standard Account Confirmation (Priority: P2)

As a newly registered creator, when my user ID and creator record are already correctly aligned, I want to confirm my email and access the welcome page without issues.

**Why this priority**: This is the happy path for new users and must continue to function perfectly.

**Independent Test**: Register a new creator via `/register`, receive the email confirmation token, and click the link to confirm the email. The test passes if the account is confirmed and the onboarding welcome page is displayed.

**Acceptance Scenarios**:

1. **Given** a new user and creator record are created with matching IDs, **When** they click the confirmation link, **Then** their email is verified and they are redirected to `/stripe-connect/welcome` successfully.

---

### User Story 3 - Add Automated Tests for Confirmation Flow (Priority: P3)

As a developer, I want to have unit and integration tests verifying both aligned and mismatched Creator/User confirmation flows, so that this issue does not regress in the future.

**Why this priority**: Automated tests ensure the durability of the fix and directly fulfill the user's requirement.

**Independent Test**: Run `go test ./...` in the repository and ensure all new tests pass successfully.

**Acceptance Scenarios**:

1. **Given** the test suite is executed, **When** testing the account confirmation and creator lookup logic, **Then** both the aligned user/creator flow and the mismatched user/creator auto-healing flow are validated and assert the correct behavior.

---

### Edge Cases

- **No Creator record exists**: If a user verifies their email but no creator record with their email exists in the database, the system should log a warning and show a user-friendly error page or redirect them to a profile creation page rather than crashing or showing a raw "Creator não encontrado" error.
- **Multiple Creator records exist with same email**: If multiple creators exist with the same email, the system should update/use the most recently updated or created record to link to the user, and log a warning for data duplication.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support looking up a creator by the user's email during the onboarding/welcome checks.
- **FR-002**: If a creator record is found with the correct email but has a mismatched `user_id`, the system MUST automatically update that creator's `user_id` to match the user's actual ID to heal the database state.
- **FR-003**: The system MUST write an info log entry when auto-healing a creator's user ID association.
- **FR-004**: The system MUST include automated tests (unit/integration) that verify the email confirmation/onboarding flows under aligned, mismatched, and missing creator scenarios.

### Key Entities

- **User**: Represents the authentication account (credentials, email verification token, verified status).
- **Creator**: Represents the creator's business/profile information (CPF, email, user_id, Stripe onboarding status).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of email confirmation requests with mismatched creator `user_id` but matching email are successfully verified and the creator record is healed automatically.
- **SC-002**: Zero HTTP 404/500 errors shown to the user on `/stripe-connect/welcome` due to mismatched creator IDs.
- **SC-003**: Automated tests exist and pass, covering all acceptance scenarios.

## Assumptions

- The user's email in the `users` table is the source of truth for the login session.
- If a creator record has a mismatched user ID but the same email, it belongs to the same user (e.g., after a database reset/seed where user IDs were shifted but emails remained the same).
