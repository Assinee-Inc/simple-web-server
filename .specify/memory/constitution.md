<!--
Sync Impact Report:
- Version change: 1.0.0
- List of modified principles: None
- Added sections: Core Principles, Technology Stack, Development Workflow, Governance
- Removed sections: None
- Templates requiring updates: None
- Follow-up TODOs: None
-->
# SimpleWebServer Constitution

## Core Principles

### I. Content Protection
Content protection is a primary feature of the system. All digital products must be protected against unauthorized distribution. This includes watermarking, download limits, and access expiration.

### II. Secure Payments
All payments must be processed through a secure, PCI-compliant payment gateway. The integration with the payment gateway must be secure and reliable, with proper handling of webhooks and transaction statuses.

### III. Data Validation
All user-provided data must be validated to ensure its integrity and correctness. This includes, but is not limited to, CPF validation, age verification, and email format validation.

### IV. Testability
The system must be thoroughly tested to ensure its quality and reliability. This includes unit tests for services, integration tests for handlers, and E2E tests for user flows. A minimum test coverage of 80% must be maintained.

### V. Modularity
The system must be designed in a modular way, with a clear separation of concerns. The project structure should follow the established pattern of handlers, services, and repositories.

## Technology Stack

The following technologies are approved for use in this project:
- **Backend**: Go (Golang)
- **Framework Web**: Chi Router
- **ORM**: GORM
- **Banco de Dados**: SQLite (desenvolvimento) / PostgreSQL (produção)
- **Frontend**: HTML + Bootstrap 5 + JavaScript
- **Pagamentos**: Stripe
- **Armazenamento**: S3 (AWS)
- **Email**: GoMail

## Development Workflow

All contributions to the project must follow this workflow:
1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Implement the changes, following the established code style and conventions.
4. Add or update tests to ensure the changes are covered.
5. Submit a pull request for review.

## Governance

This constitution is the supreme document that governs the development of this project. All development practices, code reviews, and contributions must adhere to the principles and rules outlined in this document. Amendments to this constitution require a formal proposal, review, and approval process.

**Version**: 1.0.0 | **Ratified**: 2026-04-23 | **Last Amended**: 2026-04-23