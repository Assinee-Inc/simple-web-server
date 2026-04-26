# Implementation Plan: Omit Client's Date of Birth from CSV Export

**Branch**: `feature/001-omit-birthdate-from-csv-export` | **Date**: 2026-04-23 | **Spec**: [./spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-omit-birthdate-from-csv-export/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

As a user with permission to export client data, I want to be able to export a list of clients to a CSV file without exposing sensitive information like the date of birth. This will be achieved by modifying the existing CSV export functionality to remove the `birthdate` field from the output.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: Chi v5, GORM v1.26, AWS SDK v2, Stripe Go v76, Go-Mail v0.6.2
**Storage**: SQLite (development) / PostgreSQL (production)
**Testing**: Go standard library (`testing`), `testify`, Cypress (E2E)
**Target Platform**: Linux Server (assumed)
**Project Type**: Web Service
**Performance Goals**: CSV export for 10,000 clients < 30 seconds.
**Constraints**: NEEDS CLARIFICATION
**Scale/Scope**: NEEDS CLARIFICATION

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Content Protection**: PASS (Feature enhances privacy by removing PII)
- **II. Secure Payments**: N/A
- **III. Data Validation**: N/A
- **IV. Testability**: PASS (Feature spec includes testing scenarios)
- **V. Modularity**: PASS (Assumes change can be made within existing modular structure)

## Project Structure

### Documentation (this feature)

```text
specs/001-omit-birthdate-from-csv-export/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
```text
cmd/
└── web/
    └── main.go
internal/
├── account/
├── auth/
├── config/
├── delivery/
├── library/
├── mocks/
├── sales/
├── shared/
└── subscription/
pkg/
├── cookie/
├── database/
├── gov/
├── mail/
├── middleware/
├── storage/
├── template/
└── utils/
web/
├── assets/
├── layouts/
├── mails/
├── pages/
├── partials/
└── templates/
```

**Structure Decision**: The project is a standard Go web service. The changes will be applied to the existing structure, likely within a handler and service responsible for client data and exports.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
|           |            |                                     |