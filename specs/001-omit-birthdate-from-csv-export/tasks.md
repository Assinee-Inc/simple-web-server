---
description: "Task list for Omit Client's Date of Birth from CSV Export"
---

# Tasks: Omit Client's Date of Birth from CSV Export

**Input**: Design documents from `/specs/001-omit-birthdate-from-csv-export/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1)
- Include exact file paths in descriptions

---

## Phase 1: User Story 1 - Export client data without sensitive information (Priority: P1) 🎯 MVP

**Goal**: As a user with permission to export client data, I want to be able to export a list of clients to a CSV file, so that I can use this data for external analysis, without exposing sensitive information like the date of birth.

**Independent Test**: A user can trigger the client data export and verify that the generated CSV file contains all the expected client information, except for the date of birth, by following the steps in `specs/001-omit-birthdate-from-csv-export/quickstart.md`.

### Implementation for User Story 1

- [x] T001 [US1] Modify `ClientExportCSV` function in `internal/sales/handler/client_handler.go` to remove the "Data Nascimento" header and the `client.Birthdate` field from the CSV records.
- [x] T002 [US1] Update `TestClientExportCSV_Success` test in `internal/sales/handler/client_handler_test.go` to assert that the generated CSV does not contain the "Data Nascimento" header or the birthdate data.

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 2: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup and validation.

- [x] T003 [P] Review documentation in `specs/001-omit-birthdate-from-csv-export/` to ensure it is up-to-date.
- [x] T004 Manually validate the feature by following the steps in `specs/001-omit-birthdate-from-csv-export/quickstart.md`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **User Story 1 (Phase 1)**: No dependencies - can start immediately.
- **Polish (Phase 2)**: Depends on User Story 1 completion.

### Task Dependencies

- **T002** depends on the completion of **T001**.

### Parallel Opportunities

- No parallel opportunities in this simple feature.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: User Story 1.
2. **STOP and VALIDATE**: Test User Story 1 independently.
3. Complete Phase 2: Polish.
4. Deploy/demo if ready.
