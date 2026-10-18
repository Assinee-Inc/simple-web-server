# Data Model: Fix Creator Account Confirmation

This document describes the key entities involved in the account confirmation and creator lookup flow, highlighting their fields and relationships.

## Entities

### User
Represents the authentication account.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `uint` (Primary Key) | Auto-incrementing identifier |
| `Email` | `string` | Unique email address |
| `EmailVerificationToken` | `string` | Token used to verify the email |
| `EmailVerifiedAt` | `*time.Time` | Timestamp when email was confirmed |

### Creator
Represents the creator's profile and business details.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `uint` (Primary Key) | Auto-incrementing identifier |
| `Email` | `string` | Creator's email address |
| `UserID` | `uint` | Foreign key linking to `users.id` |
| `CPF` | `string` | Tax identification number |

## Relationships

- **One-to-One**: A `User` is associated with at most one `Creator` profile.
  - Linked via `creators.user_id = users.id`.
- **Constraint / Invariance**: The email of a `User` and its associated `Creator` must be identical.
  - If a creator profile exists for an email `E`, it must map to the user profile with email `E` (i.e. `creator.user_id` must equal `user.id`).

## State Validation and Healing Logic

When querying the creator by user email:
1. Try fetching via JOIN: `creators.user_id = users.id AND users.email = ?`.
2. Fallback to direct query on `creators` and `users` tables if the JOIN fails.
3. If both exist but have mismatched IDs, update `creators.user_id = users.id` and save to resolve the mismatch.
