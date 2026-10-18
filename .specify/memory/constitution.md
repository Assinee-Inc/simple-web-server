<!--
Sync Impact Report:
- Version change: 1.0.0 → 1.1.0
- List of modified principles:
  - I. Content Protection → I. Content Protection (strengthened with declarative MUST rules)
  - II. Secure Payments → II. Secure Payments (strengthened with declarative MUST rules; includes Stripe and Woovi/Pix)
  - III. Data Validation → III. Data Validation (strengthened with declarative MUST rules)
  - IV. Testability → IV. Testability (strengthened with declarative MUST rules)
  - V. Modularity → V. Modularity (expanded with domain-module architecture and primitive ID cross-module boundary rules)
- Added sections: None
- Removed sections: None
- Templates requiring updates:
  - ✅ .specify/templates/plan-template.md (verified aligned)
  - ✅ .specify/templates/spec-template.md (verified aligned)
  - ✅ .specify/templates/tasks-template.md (verified aligned)
  - ✅ README.md (updated legacy SQLite development references to PostgreSQL/Docker)
- Follow-up TODOs: None
-->
# SimpleWebServer Constitution

## Core Principles

### I. Content Protection
Content protection is a primary feature of the system. All digital products MUST be protected against unauthorized distribution. Every digital delivery flow MUST enforce dynamic PDF watermarking with buyer identification, strict download limits, and time-bound access expiration.

### II. Secure Payments
All payments MUST be processed through secure, compliant payment gateways (Stripe and Woovi/Pix). Payment integrations MUST verify webhook signatures, enforce idempotency against duplicate transaction events, and maintain accurate transaction status lifecycles before granting product access.

### III. Data Validation
All user-provided data MUST be validated on the server side to ensure integrity and correctness before persistence or external API calls. This MUST include CPF checksum validation, age verification, email format verification, and sanitization against injection and XSS attacks.

### IV. Testability
The system MUST be thoroughly tested to ensure quality and reliability across all modules. Contributions MUST include unit tests for business services, integration tests for HTTP handlers and repositories, and end-to-end tests for critical user flows. A minimum test coverage of 80% MUST be maintained for core business logic.

### V. Modularity
The system MUST be organized into self-contained business domain modules (`auth`, `account`, `library`, `sales`, `delivery`, `subscription`), each encapsulating its own `model`, `repository`, `service`, and `handler` layers. Cross-module communication MUST occur strictly via primitive IDs (`uint`) rather than direct domain struct imports, and input DTOs MUST reside in leaf `model/` packages to prevent circular dependencies.

## Technology Stack

The following technologies are approved for use in this project:
- **Backend**: Go (Golang)
- **Framework Web**: Chi Router
- **ORM**: GORM
- **Banco de Dados**: PostgreSQL (desenvolvimento via Docker e produção) / SQLite (suporte legado/testes isolados)
- **Frontend**: HTML Templates + DaisyUI 4 + Tailwind CSS 3 + Vanilla JavaScript (Bootstrap 5 apenas em layouts legados em migração)
- **Pagamentos**: Stripe & Woovi (OpenPix)
- **Armazenamento**: AWS S3
- **Email**: GoMail (SMTP)

## Development Workflow

All contributions to the project MUST follow this workflow:
1. Create a feature or bugfix branch from `develop` following the project naming convention.
2. Implement the changes adhering to the modular architecture (`handler` → `service` → `repository`) and Go code conventions.
3. Add or update unit and integration tests ensuring all test suites pass (`make test`).
4. Run security and configuration checks (`make security-check`) when touching authentication, payments, or middleware.
5. Submit a pull request for code review and verification against this Constitution.

## Governance

This constitution is the supreme document governing the architecture and development of SimpleWebServer. All specifications, implementation plans, code reviews, and contributions MUST verify compliance with these principles.

- **Amendment Procedure**: Any proposed change to principles, technology stack, or governance MUST be documented with a clear rationale, reviewed by project maintainers, and synchronized across dependent `.specify` templates and project documentation.
- **Versioning Policy**: This constitution follows Semantic Versioning (`MAJOR.MINOR.PATCH`):
  - `MAJOR`: Backward-incompatible principle removals or fundamental architectural redefinitions.
  - `MINOR`: Addition of new principles/sections or material expansion of existing architectural guidance.
  - `PATCH`: Wording clarifications, typo fixes, or non-semantic refinements.
- **Compliance Review**: Every feature implementation plan (`plan.md`) MUST include a `Constitution Check` gate validating adherence before implementation begins.

**Version**: 1.1.0 | **Ratified**: 2026-04-23 | **Last Amended**: 2026-10-18