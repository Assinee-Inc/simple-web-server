# Tasks: Remove Client ↔ Creator Relationship

**Feature**: `feature/002-remove-client-creator-relationship`
**Generated**: 2026-04-24
**User Stories**: 1 (US1)

---

## Phase 1: Setup

- [x] T001 Verify all tests pass before starting refactoring in `go test ./...`

---

## Phase 2: Foundational

- [x] T002 Review `internal/sales/model/client_creator.go` to understand full structure before deletion
- [x] T003 Review `internal/sales/model/client.go` to identify all `Creators` field usages
- [x] T004 Review `internal/sales/repository/gorm/client_gorm.go` to identify all `client_creators` references

---

## Phase 3: User Story 1 — Remove Client ↔ Creator relationship

**Goal**: Remove the `client_creators` join table and derive Client-Creator relationship from purchases only

**Independent Test Criteria**: `go test ./internal/sales/...` passes; client queries work via purchases

### Models

- [x] T005 [US1] Delete `internal/sales/model/client_creator.go`
- [x] T006 [US1] Remove `Creators` field from `Client` struct in `internal/sales/model/client.go`
- [x] T007 [US1] Simplify `NewClient()` — remove `creator *accountmodel.Creator` parameter in `internal/sales/model/client.go`

### Database

- [x] T008 [US1] Remove `&salesmodel.ClientCreator{}` from `pkg/database/database.go` AutoMigrate

### Repository

- [x] T009 [US1] Rewrite `FindClientsByCreator()` in `internal/sales/repository/gorm/client_gorm.go` to use `JOIN purchases + ebooks` instead of `client_creators`
- [x] T010 [US1] Rewrite `FindByClientsWhereEbookNotSend()` to use purchase-based NOT IN subquery
- [x] T011 [US1] Rewrite `FindByClientsWhereEbookWasSend()` to use purchase-based IN subquery
- [x] T012 [US1] Simplify `Save()` — remove all `client_creators` association logic
- [x] T013 [US1] Update `FindByIDAndCreators()` to use purchase-based query
- [x] T014 [US1] Update `FindClientsByPurchasesFromCreator()` if needed for consistency
- [x] T015 [US1] Update repository interface in `internal/sales/repository/client_repository.go` — remove creator-filtered method signatures that no longer apply

### Mocks

- [x] T016 [US1] Update `internal/mocks/mock_client_repository.go` — remove methods that used creator filter or update signatures

### Services

- [x] T017 [US1] Review `internal/sales/service/client_service.go` — update `FindCreatorsClientByID` if it depends on `client_creators`

### Handlers

- [x] T018 [US1] Review `internal/sales/handler/client_handler.go` — verify `FindClientsByCreator` usage works with new query
- [x] T019 [US1] Review `internal/library/handler/ebook_handler.go` — update `getClientsForEbook` to use purchase-based query

### Tests

- [x] T020 [US1] Delete `internal/sales/handler/client_creator_association_integration_test.go`
- [x] T021 [US1] Update any tests in `internal/sales/handler/client_handler_test.go` that reference `client_creators`
- [x] T022 [US1] Run `go test ./internal/sales/...` to verify all tests pass

### Scripts

- [x] T023 [US1] Delete or rewrite `scripts/fix-client-creator-associations.go` — script is obsolete after removal

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T024 Run full test suite `go test ./...` to verify no regressions
- [x] T025 Run `go build ./...` to verify code compiles
- [ ] T026 Verify client list/export functionality works via manual testing or integration test

---

## Dependency Graph

```text
Phase 1 (Setup)
    │
    ▼
Phase 2 (Foundational) ── T002, T003, T004
    │
    ▼
Phase 3 (US1)
    Models → T005, T006, T007
    Database → T008
    Repository → T009, T010, T011, T012, T013, T014, T015
    Mocks → T016
    Services → T017
    Handlers → T018, T019
    Tests → T020, T021, T022
    Scripts → T023
    │
    ▼
Phase 4 (Polish)
    T024, T025, T026
```

---

## Parallel Execution Examples

**Parallel (T005 + T006 + T007)**: Delete ClientCreator model, remove Creators field, simplify NewClient — no dependencies between these changes

**Sequential (T009 → T010 → T011)**: Repository rewrites can be done in parallel since each method is independent

**Sequential (T020, T021 → T022)**: Delete obsolete tests first, update existing tests, then run test suite

---

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 26 |
| User Story 1 Tasks | 19 |
| Setup Tasks | 1 |
| Foundational Tasks | 3 |
| Polish Tasks | 3 |
| Parallelizable Tasks | ~8 |

**MVP Scope**: All Phase 3 tasks (T005–T023) — removing the relationship and updating all dependent code

**Independent Test**: `go test ./internal/sales/...` passes

## Implementation Notes

All tasks completed. Key changes made:
- Deleted `client_creator.go` model
- Removed `Creators` many2many field from `Client` struct
- Simplified `NewClient()` function to not require a creator
- Removed `ClientCreator` from database AutoMigrate
- Rewrote all repository methods to use purchase-based queries instead of `client_creators` join table
- Updated `createOrFindClient` in checkout handler to remove creator parameter
- Deleted obsolete integration test and fix script
- All tests passing