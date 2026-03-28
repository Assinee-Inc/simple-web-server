const { defineConfig } = require('cypress')

module.exports = defineConfig({
  e2e: {
    // Must match the Go server address (make run → :8080)
    baseUrl: 'http://localhost:8080',

    specPattern: 'cypress/e2e/**/*.cy.js',

    // Support file loaded before every spec
    supportFile: 'cypress/support/e2e.js',

    // Keep videos off by default; enable in CI by overriding with --env
    video: false,

    // Screenshots only on failure — kept in .gitignore
    screenshotOnRunFailure: true,

    // Viewport that matches the checkout card layout (max-w-2xl)
    viewportWidth: 1280,
    viewportHeight: 900,

    // How long to wait for a command before failing (ms)
    defaultCommandTimeout: 8000,

    // How long to wait for a page load (ms)
    pageLoadTimeout: 30000,

    setupNodeEvents(on, config) {
      // Future: add tasks here (e.g. seeding test DB, reading fixtures from Go side)
      return config
    },
  },

  env: {
    // Override on CI: CYPRESS_BASE_URL=https://staging.example.com npx cypress run
    // These are not secrets — real secrets go in cypress.env.json (git-ignored)

    // A valid ebook PublicID that exists in your dev DB.
    // Override per run: CYPRESS_EBOOK_ID=ebk_abc123 npx cypress run
    EBOOK_ID: 'ebk_CHANGE_ME',
  },
})
