// -----------------------------------------------------------------------------
// Custom Cypress commands for the checkout flow
//
// Naming convention mirrors the JS in purchase.checkout.js so that if the
// field IDs change there, they are easy to find here.
// -----------------------------------------------------------------------------

/**
 * cy.fillCheckoutForm(data)
 *
 * Fills every visible input on #checkoutForm.
 * Accepts a plain object; any field omitted will be skipped (useful for
 * partial-fill tests that verify the pay button stays disabled).
 *
 * @param {object} data
 * @param {string} [data.name]       Full name (min 3 chars)
 * @param {string} [data.cpf]        CPF with mask, e.g. "123.456.789-09"
 * @param {string} [data.birthdate]  DD/MM/AAAA
 * @param {string} [data.email]      Valid e-mail address
 * @param {string} [data.phone]      Phone with mask, e.g. "(11) 9 1234-5678"
 */
Cypress.Commands.add('fillCheckoutForm', (data = {}) => {
  if (data.name !== undefined) {
    cy.get('#name').clear().type(data.name)
  }
  if (data.cpf !== undefined) {
    // The JS mask expects the raw value typed; we type the masked form
    // because the input fires 'input' events which update validation state.
    cy.get('#cpf').clear().type(data.cpf)
  }
  if (data.birthdate !== undefined) {
    cy.get('#birthdate').clear().type(data.birthdate)
  }
  if (data.email !== undefined) {
    cy.get('#email').clear().type(data.email)
  }
  if (data.phone !== undefined) {
    cy.get('#phone').clear().type(data.phone)
  }
})

/**
 * cy.mockValidateCustomer(response, statusCode)
 *
 * Stubs POST /api/validate-customer.
 * Call this BEFORE cy.get('#checkoutForm') submit so the intercept is
 * registered before the network request fires.
 *
 * @param {object} response    JSON body the stub returns
 * @param {number} statusCode  HTTP status (default 200)
 *
 * The alias @validateCustomer is always set so specs can
 * cy.wait('@validateCustomer') to synchronise assertions.
 */
Cypress.Commands.add('mockValidateCustomer', (response, statusCode = 200) => {
  cy.intercept('POST', '/api/validate-customer', {
    statusCode,
    body: response,
    headers: { 'Content-Type': 'application/json' },
  }).as('validateCustomer')
})

/**
 * cy.mockCreateCheckout(response, statusCode)
 *
 * Stubs POST /api/create-ebook-checkout.
 * In happy-path tests the response must contain a `url` field so that
 * purchase.checkout.js calls window.location.href = response.url.
 * We intercept that redirect separately to avoid leaving the app.
 *
 * @param {object} response    JSON body the stub returns
 * @param {number} statusCode  HTTP status (default 200)
 */
Cypress.Commands.add('mockCreateCheckout', (response, statusCode = 200) => {
  cy.intercept('POST', '/api/create-ebook-checkout', {
    statusCode,
    body: response,
    headers: { 'Content-Type': 'application/json' },
  }).as('createCheckout')
})

/**
 * cy.visitCheckout(ebookId)
 *
 * Navigates to /checkout/:ebookId and waits for the form to be visible.
 * Reads ebookId from Cypress.env('EBOOK_ID') if not provided.
 *
 * @param {string} [ebookId]
 */
Cypress.Commands.add('visitCheckout', (ebookId) => {
  const id = ebookId || Cypress.env('EBOOK_ID')
  cy.visit(`/checkout/${id}`)
  cy.get('#checkoutForm').should('be.visible')
})

/**
 * cy.csrfToken()
 *
 * Yields the CSRF token string embedded in the hidden #csrfToken input.
 * The checkout page injects it server-side via {{.CSRFToken}}.
 * On public routes with no active session the value may be empty — that is
 * acceptable because ValidateCustomer and CreateEbookCheckout do not
 * server-validate the token on these public endpoints (confirmed in
 * checkout_handler.go: the csrfToken field is read but never checked).
 *
 * Use this command if a future test needs to assert the token is present
 * or to forward it explicitly.
 */
Cypress.Commands.add('csrfToken', () => {
  return cy.get('#csrfToken').invoke('val')
})
