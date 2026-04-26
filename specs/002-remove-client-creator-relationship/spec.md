# Feature Specification: Remove Client ↔ Creator Relationship

**Feature Branch**: `feature/002-remove-client-creator-relationship`
**Created**: 2026-04-24
**Status**: Draft
**Input**: User request to remove the Client-Creator relationship, deriving it from purchases instead.

## Summary

Remove the explicit `client_creators` join table and `Creators` field on the `Client` model. A Client will only be associated with a Creator through their purchases (Client → Purchase → Ebook → Creator). This eliminates data redundancy and ensures the relationship is always derived from the source of truth.

## User Scenarios & Testing

### User Story 1 - Clients are associated to creators only via purchases

As a developer, I want the Client-Creator relationship to be derived from purchases, so that the data model is simpler and doesn't allow inconsistent state (e.g., a client associated with a creator but no purchase exists).

**Acceptance Scenarios**:
1. **Given** a client has purchased an ebook from a creator, **When** we query for the client's creators, **Then** the creator is returned via the purchase chain.
2. **Given** a client has no purchases from a creator, **When** we query for the client's creators, **Then** the creator is NOT returned.
3. **Given** the system has no `client_creators` table, **When** existing code queries clients by creator, **Then** it uses `purchases` + `ebooks` tables instead.

## Requirements

### Functional Requirements

- **FR-001**: The `client_creators` join table must be removed from the database schema.
- **FR-002**: The `Creators` field must be removed from the `Client` model.
- **FR-003**: `Client.FindClientsByCreator` must be rewritten to query via `purchases` → `ebooks` tables instead of `client_creators`.
- **FR-004**: `Client.FindByClientsWhereEbookNotSend` must use purchase-based queries.
- **FR-005**: `Client.FindByClientsWhereEbookWasSend` must use purchase-based queries.
- **FR-006**: `Client.Save` must not reference or create `client_creators` associations.
- **FR-007**: All tests referencing `client_creators` must be updated or removed.
- **FR-008**: The `ClientCreator` model must be deleted.

### Non-Functional Requirements

- **NFR-001**: No performance degradation for client queries — purchase-based queries must be equally efficient.
- **NFR-002**: No data loss — existing client data must be preserved during migration.

## Key Entities

### Client (modified)

**Removed fields**:
- `Creators []*accountmodel.Creator` (many2many via client_creators)

**Retained fields**:
- `ID`, `PublicID`, `Name`, `CPF`, `Birthdate`, `Email`, `Phone`, `Validated`, `Purchases`

### ClientCreator (deleted)

Entire model and table removed.

### Purchase (unchanged)

- `ClientID`, `EbookID` — already links Client to Creator via Ebook.CreatorID

### Ebook (unchanged)

- `CreatorID` — already links to Creator

## Technical Approach

### Backend Changes

1. Delete `internal/sales/model/client_creator.go`
2. Remove `Creators` field from `Client` model in `client.go`
3. Simplify `NewClient()` — remove `creator` parameter
4. Rewrite `FindClientsByCreator` to use:
   ```sql
   SELECT DISTINCT clients.* FROM clients
   JOIN purchases ON purchases.client_id = clients.id
   JOIN ebooks ON ebooks.id = purchases.ebook_id
   WHERE ebooks.creator_id = ?
   ```
5. Rewrite `FindByClientsWhereEbookNotSend` similarly
6. Rewrite `FindByClientsWhereEbookWasSend` similarly
7. Simplify `Save()` — remove all `client_creators` logic
8. Update repository interface in `client_repository.go`
9. Update mocks in `mock_client_repository.go`
10. Delete `client_creator_association_integration_test.go`
11. Update other tests that reference `client_creators`
12. Remove `&salesmodel.ClientCreator{}` from `pkg/database/database.go`

### Frontend Impact Analysis

**No HTML templates directly reference `client_creators` or the `Creators` field.**

Templates using Client data do not display the Client-Creator relationship:

| Template | Client Usage | Changes Needed |
|----------|--------------|----------------|
| `web/pages/client/list.html` | Lists clients (Name, Email, Phone) | None |
| `web/pages/client/update.html` | Edit client form | None |
| `web/pages/ebook/send.html` | Lists clients for ebook delivery | None |

The refactoring is **purely a backend/data model change**. All client-facing templates work with Client fields only and require no modifications.

## Success Criteria

- **SC-001**: All client queries that previously used `client_creators` now use `purchases` + `ebooks` tables.
- **SC-002**: No code references `client_creators` table or `ClientCreator` model.
- **SC-003**: All tests pass with the new query approach.
- **SC-004**: Client data is preserved (no purchases lost).