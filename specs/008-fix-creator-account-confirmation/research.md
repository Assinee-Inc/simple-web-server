# Research: Fix Creator Account Confirmation Mismatch

This research analyzes the root cause of the failure in the creator account confirmation flow and details the selected solution.

## Problem Analysis

The user logs show the following query failure during the account confirmation flow redirect (`/stripe-connect/welcome`):

```sql
SELECT "creators".* FROM "creators"
JOIN users ON users.id = creators.user_id
WHERE users.email = 'glauciajuliete@outlook.com' AND "creators"."deleted_at" IS NULL
ORDER BY "creators"."id" LIMIT 1
```

This query returned 0 rows (`record not found`).
An inspection of the database revealed:
1. `users` table has a user record with ID `34` and email `glauciajuliete@outlook.com`.
2. `creators` table has a creator record with ID `1`, email `glauciajuliete@outlook.com`, and `user_id = 1`.
3. No user with ID `1` exists in the database.

Because the creator record points to a non-existent `user_id = 1`, the JOIN query fails to find the matching creator for the logged-in user `34`, even though their emails match.

### Mismatch Cause
This mismatch occurs when:
1. The database is seeded or reset (e.g. truncated), wiping/re-creating users table records and assigning new auto-increment IDs (such as `34`).
2. The `creators` table is not correctly synchronized or is restored/persistent with old user IDs.
3. A user registers/confirms their email under an email that already exists in the `creators` table but was assigned to an older, deleted user ID.

## Proposed Solution: Auto-Healing Database Linkage

Rather than hard-failing with "Creator não encontrado" or HTTP 404 when `FindCreatorByUserEmail` fails due to `RecordNotFound`, the system can detect this state and automatically heal it:

1. **Query Join**: Attempt the standard JOIN query: `SELECT * FROM creators JOIN users ON users.id = creators.user_id WHERE users.email = ?`.
2. **Fallback Email Query**: If no record is found:
   - Query the `creators` table directly by `email`.
   - Query the `users` table directly by `email`.
   - If both records exist but `creator.user_id != user.id`:
     - Update the creator's `user_id` to match the user's `id`.
     - Save/update the creator record in the database.
     - Return the healed creator record.
3. **Graceful Degradation**: If no creator record exists at all for the email, return a clear error (or redirect to a signup/onboarding screen) instead of panicking/crashing.

## Alternatives Considered

### Alternative 1: Enforce Unique Constraints on Emails and Roll Back Registrations
Enforce a unique constraint on the `email` column in the `creators` table.
- **Why rejected**: A database schema migration is risky for existing production data. Additionally, it does not heal existing mismatched states that already exist in local dev environments or testing.

### Alternative 2: Fail and Prompt User to Re-register
If the creator is not found, tell the user to sign up again.
- **Why rejected**: Extremely poor user experience. The user already filled out their registration details (CPF, phone, name), which are stored in the database. Telling them to register again would fail anyway because their CPF or email already exists.
