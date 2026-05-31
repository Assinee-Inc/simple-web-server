# Feature Specification: Woovi Checkout Redirection

**Feature Branch**: `[006-woovi-checkout-redirect]`  
**Created**: 2026-05-31  
**Status**: Draft  
**Input**: User description: "A página de checkout está direcionando para o stripe e deveria redirecionar para o Woovi nesta nova modificacao que estamos trabalhando (arquivo modificados na branch atual)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Complete Payment via Woovi (Priority: P1)

As a customer, when I finish selecting my products and proceed to checkout, I want to be redirected to the Woovi payment platform so that I can complete my purchase using a modern and efficient payment method.

**Why this priority**: This is the core functionality of the feature. Without it, the user cannot pay, and the business cannot generate revenue from sales using the new payment gateway.

**Independent Test**: This can be tested by creating an order and proceeding to the checkout page. The test passes if the user is correctly redirected to the Woovi payment page.

**Acceptance Scenarios**:

1. **Given** a user has items in their cart and is on the checkout page, **When** they click the "Pay" or "Complete Order" button, **Then** they are redirected to the Woovi payment gateway.
2. **Given** a user is on the checkout page, **When** the system initiates the payment process, **Then** the existing redirection to Stripe is no longer triggered.
3. **Given** a user is redirected to Woovi, **When** they complete the payment successfully, **Then** they are redirected back to a confirmation page on our site.
4. **Given** a user is redirected to Woovi, **When** they cancel the payment, **Then** they are redirected back to the checkout or cart page on our site.

### Edge Cases

- What happens if the Woovi service is unavailable? The user should be shown an error message and should not be left on a broken page.
- What happens if the order details (e.g., total amount) are invalid when sent to Woovi? The system should handle the error gracefully and log the issue for investigation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST redirect the user to the Woovi payment gateway when initiating the checkout process.
- **FR-002**: The system MUST stop redirecting users to the Stripe payment gateway.
- **FR-003**: The system MUST send all necessary order information (e.g., total amount, order ID) to Woovi to create the payment charge.
- **FR-004**: The system MUST have a configurable mechanism to enable or disable the Woovi redirection, preferably via environment variables.
- **FR-005**: The system MUST handle success, failure, and cancellation callbacks or webhooks from Woovi to update the order status accordingly.

### Key Entities

- **Order**: Represents the customer's purchase, containing items, total amount, and status (e.g., pending, paid, cancelled).
- **Payment**: Represents the financial transaction, linked to an Order. It includes details like the payment provider (Woovi), transaction ID, and status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of checkout initiations for the new payment flow redirect users to the Woovi platform.
- **SC-002**: 0% of checkout initiations for the new payment flow redirect users to the Stripe platform.
- **SC-003**: 99% of users who are redirected to Woovi experience a successful page load of the Woovi payment interface within 5 seconds.
- **SC-004**: The post-payment order status (e.g., "paid", "cancelled") is correctly updated for at least 99.5% of transactions initiated via Woovi.

## Assumptions

- The necessary API keys, secrets, and other credentials for Woovi integration are available and configured securely in the application's environment.
- The project already has a mechanism for handling external payment gateway callbacks or webhooks, which can be adapted for Woovi.
- The decision of which payment gateway to use (Stripe vs. Woovi) is determined by the current feature branch and does not require a dynamic, user-facing selection mechanism at this time.
