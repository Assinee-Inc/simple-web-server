// -----------------------------------------------------------------------------
// Checkout flow — E2E spec
//
// These tests cover the client-side behaviour of purchase/checkout.html and
// purchase.checkout.js. Both API calls are stubbed so no live DB or Stripe
// key is required.
//
// Pre-condition: the Go server must be running (make run) and
// CYPRESS_EBOOK_ID must point to a valid, published ebook in the dev DB.
// -----------------------------------------------------------------------------

describe('Checkout page', () => {
  let customer
  let fixtures

  before(() => {
    cy.fixture('checkout').then((data) => {
      fixtures = data
      customer = data.validCustomer
    })
  })

  // ---------------------------------------------------------------------------
  // Page load
  // ---------------------------------------------------------------------------
  describe('Initial state', () => {
    beforeEach(() => {
      cy.visitCheckout()
    })

    it('renders the ebook title and price', () => {
      cy.get('.card-body').should('be.visible')
      // The product summary block contains at least the title text
      cy.get('.bg-base-200 .font-semibold').should('not.be.empty')
    })

    it('pay button is disabled on load', () => {
      cy.get('#payButton').should('be.disabled')
    })

    it('has a hidden CSRF token input', () => {
      // Token may be empty on public routes (server does not validate it
      // for these endpoints) but the element must be present.
      cy.get('#csrfToken').should('exist')
    })

    it('shows the back link to the sales page', () => {
      cy.get('a[href*="/sales/"]').should('be.visible')
    })
  })

  // ---------------------------------------------------------------------------
  // Form validation — pay button enable/disable
  // ---------------------------------------------------------------------------
  describe('Client-side form validation', () => {
    beforeEach(() => {
      cy.visitCheckout()
    })

    it('enables pay button only when all fields are valid', () => {
      cy.get('#payButton').should('be.disabled')
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').should('not.be.disabled')
    })

    it('keeps pay button disabled if name is too short', () => {
      cy.fillCheckoutForm({ ...customer, name: 'AB' })
      cy.get('#payButton').should('be.disabled')
    })

    it('keeps pay button disabled if CPF has fewer than 11 digits', () => {
      // Provide only 10 digits (without mask separators)
      cy.fillCheckoutForm({ ...customer, cpf: '1234567890' })
      cy.get('#payButton').should('be.disabled')
    })

    it('keeps pay button disabled if email has no @', () => {
      cy.fillCheckoutForm({ ...customer, email: 'notanemail' })
      cy.get('#payButton').should('be.disabled')
    })

    it('keeps pay button disabled if phone mask length is not 16 chars', () => {
      // "(11) 9 1234-567" = 15 chars — one digit short
      cy.fillCheckoutForm({ ...customer, phone: '(11) 9 1234-567' })
      cy.get('#payButton').should('be.disabled')
    })

    it('keeps pay button disabled if birthdate is incomplete', () => {
      cy.fillCheckoutForm({ ...customer, birthdate: '15/04/19' })
      cy.get('#payButton').should('be.disabled')
    })
  })

  // ---------------------------------------------------------------------------
  // Happy path: validate → create checkout → redirect
  // ---------------------------------------------------------------------------
  describe('Happy path', () => {
    beforeEach(() => {
      cy.visitCheckout()
      cy.mockValidateCustomer(fixtures.validateCustomerSuccess)
      // URL fixture aponta para "/" (home pública) — carrega instantaneamente,
      // evitando o pageLoadTimeout causado por chamadas ao Stripe com ID falso.
      // cy.intercept NÃO intercepta window.location.href, apenas fetch/XHR.
      cy.mockCreateCheckout(fixtures.createCheckoutSuccess)
    })

    it('calls validate-customer with the correct payload', () => {
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()

      cy.wait('@validateCustomer').then((interception) => {
        const body = interception.request.body
        // CPF must be digits-only (purchase.checkout.js strips non-digits)
        expect(body.cpf).to.match(/^\d{11}$/)
        // Phone must be digits-only
        expect(body.phone).to.match(/^\d{10,11}$/)
        expect(body.name.length).to.be.gte(3)
        expect(body.email).to.include('@')
        expect(body.ebookId).to.equal(Cypress.env('EBOOK_ID'))
      })
    })

    it('calls create-ebook-checkout after successful validation', () => {
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()

      cy.wait('@validateCustomer')
      cy.wait('@createCheckout').then((interception) => {
        expect(interception.request.body.ebookId).to.equal(Cypress.env('EBOOK_ID'))
      })
    })

    it('shows loading spinner while waiting for API', () => {
      // Delay the stub so the spinner is visible
      cy.intercept('POST', '/api/validate-customer', (req) => {
        req.reply({ delay: 500, body: fixtures.validateCustomerSuccess })
      }).as('validateCustomerSlow')

      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()

      cy.get('#loadingSpinner').should('be.visible')
      cy.wait('@validateCustomerSlow')
    })
  })

  // ---------------------------------------------------------------------------
  // Already purchased
  // ---------------------------------------------------------------------------
  describe('Already purchased', () => {
    beforeEach(() => {
      cy.visitCheckout()
      cy.mockValidateCustomer(fixtures.validateCustomerAlreadyPurchased, 409)
    })

    it('hides the form and shows the already-purchased alert', () => {
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()
      cy.wait('@validateCustomer')

      cy.get('#checkoutForm').should('not.be.visible')
      cy.get('#alreadyPurchasedMessage').should('be.visible')
    })

    it('shows the creator email as a mailto link', () => {
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()
      cy.wait('@validateCustomer')

      cy.get('#creatorContact a[href^="mailto:"]')
        .should('be.visible')
        .and('contain', fixtures.validateCustomerAlreadyPurchased.creator_email)
    })
  })

  // ---------------------------------------------------------------------------
  // Validation failure (RF rejection)
  // ---------------------------------------------------------------------------
  describe('Validation failure', () => {
    beforeEach(() => {
      cy.visitCheckout()
      cy.mockValidateCustomer(fixtures.validateCustomerError, 400)
    })

    it('shows an alert with the server error message', () => {
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()
      cy.wait('@validateCustomer')

      // purchase.checkout.js calls alert() — stub it to capture the message
      cy.on('window:alert', (msg) => {
        expect(msg).to.include(fixtures.validateCustomerError.error)
      })
    })

    it('re-enables the pay button after validation error', () => {
      cy.fillCheckoutForm(customer)
      cy.get('#payButton').click()
      cy.wait('@validateCustomer')

      cy.get('#payButton').should('not.be.disabled')
    })
  })

  // ---------------------------------------------------------------------------
  // create-ebook-checkout failure
  // ---------------------------------------------------------------------------
  describe('Checkout creation failure', () => {
    beforeEach(() => {
      cy.visitCheckout()
      cy.mockValidateCustomer(fixtures.validateCustomerSuccess)
      cy.mockCreateCheckout(fixtures.createCheckoutError, 500)
    })

    it('shows an alert when Stripe session creation fails', () => {
      cy.fillCheckoutForm(customer)

      cy.on('window:alert', (msg) => {
        expect(msg).to.include('Erro ao criar sessão de pagamento')
      })

      cy.get('#payButton').click()
      cy.wait('@validateCustomer')
      cy.wait('@createCheckout')
    })
  })
})
