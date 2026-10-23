# Implementation Plan: Reduce Presigned URL Expiration to 60 Seconds

**Branch**: `011-reduce-presigned-url-expiration` | **Date**: 2026-10-23 | **Spec**: [specs/011-reduce-presigned-url-expiration/spec.md](spec.md)  
**Input**: Feature specification from `specs/011-reduce-presigned-url-expiration/spec.md`

## Summary

Reduce the time-to-live (TTL) of AWS S3 presigned URLs (for direct file uploads, downloads, and previews) to strictly 60 seconds (down from 15 minutes / 5 minutes). This closes potential interception windows and reinforces Constitution Principle I (Content Protection). Client application retains form context and provides clear feedback with single-click retry if a link expires before transfer initiation.

## Technical Context

**Language/Version**: Go 1.23+ (backend), TypeScript 5.6+ / React 18 (frontend)  
**Primary Dependencies**: `github.com/aws/aws-sdk-go-v2/service/s3` v1.78.1, Chi Router v5, GORM v1.25, Tailwind CSS / DaisyUI  
**Storage**: AWS S3 (object storage), PostgreSQL (relational database)  
**Testing**: Go standard testing package (`go test ./...`), `testify/assert`, `testify/mock`  
**Target Platform**: Linux server, modern web browsers  
**Project Type**: Fullstack web service (Go backend + React frontend)  
**Performance Goals**: URL generation latency < 50ms, upload handshake initiation within 60s, retry URL generation < 100ms  
**Constraints**: Ephemeral presigned URL expiration strictly equal to 60 seconds; S3 stream connection established prior to t=60s must continue uninterrupted to completion  
**Scale/Scope**: Impacts `internal/library/service`, `pkg/storage`, frontend upload/library pages, and unit test suites

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evaluation |
|---|---|---|
| **I. Content Protection** | **PASSED** | Directly enhances content protection by drastically reducing unauthorized URL reuse window from 15m/5m down to 60 seconds. |
| **II. Secure Payments** | **N/A** | Does not modify payment gateways, webhooks, or financial ledger logic. |
| **III. Data Validation** | **PASSED** | Preserves all existing file metadata and namespace security checks (`files/{creator_id}/`). |
| **IV. Testability** | **PASSED** | Comprehensive unit tests for `FileService`, `S3Storage`, and handlers will be updated/added to assert `ExpiresIn: 60` and `60 * time.Second`. |
| **V. Modularity** | **PASSED** | Changes are encapsulated within the `library` module and `pkg/storage` infrastructure layer, keeping clear separation of concerns. |

## Project Structure

### Documentation (this feature)

```text
specs/011-reduce-presigned-url-expiration/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── file-upload-url.json
│   └── file-library.json
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/
└── library/
    ├── handler/
    │   ├── file_handler.go           # Presign upload & complete endpoints
    │   └── file_handler_test.go      # Handler test asserting expires_in 60
    └── service/
        ├── file_service.go           # Set expirationSeconds = 60, expiresTime = 60s
        └── file_service_test.go      # Unit tests asserting 60s TTL and DTO response

pkg/
└── storage/
    ├── s3.go                         # Ensure s3 presigner passes 60s TTL to AWS SDK
    └── s3_test.go                    # Storage unit tests

frontend/
└── src/
    ├── lib/
    │   └── s3-upload.ts              # Error handling & expiration detection for PUT
    └── pages/
        └── files/
            └── FileLibraryPage.tsx   # Upload form retry handling without losing file state
```

**Structure Decision**: Standard Docffy modular web architecture. Backend changes are confined to `internal/library` and `pkg/storage`. Frontend changes are confined to `frontend/src/pages/files` and `frontend/src/lib`.

## Complexity Tracking

*No violations of project principles or architecture rules. Complexity is minimal (configuration of TTL values and frontend error message clarity).*
