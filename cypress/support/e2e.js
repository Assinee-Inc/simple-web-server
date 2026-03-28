// -----------------------------------------------------------------------------
// Global E2E setup — loaded once before every spec file.
// -----------------------------------------------------------------------------

// Import custom commands
import './commands'

// ---------------------------------------------------------------------------
// Exception handling
// ---------------------------------------------------------------------------
// Suppress uncaught exceptions that originate in third-party scripts
// (e.g. Stripe.js loaded on the real checkout page).
// Remove this if you want Cypress to fail on any JS error.
Cypress.on('uncaught:exception', (err) => {
  // Allow the test to continue; the error is still logged to the console.
  // If you want to fail on specific errors, inspect err.message here and
  // return true for those cases.
  return false
})

// ---------------------------------------------------------------------------
// Global before / beforeEach hooks
// ---------------------------------------------------------------------------
beforeEach(() => {
  // Clear browser storage between tests to avoid state leakage.
  // Gorilla sessions are cookie-based; cy.clearCookies() handles that.
  cy.clearCookies()
  cy.clearLocalStorage()
})
