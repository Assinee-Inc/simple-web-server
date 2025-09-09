package config

// StripeOnboardingConfig defines which routes require Stripe onboarding
type StripeOnboardingConfig struct {
	// Routes that are completely excluded from onboarding checks
	ExcludedPaths []string

	// Routes that require onboarding to be completed
	ProtectedPaths []string
}

// GetStripeOnboardingConfig returns the configuration for Stripe onboarding middleware
func GetStripeOnboardingConfig() *StripeOnboardingConfig {
	return &StripeOnboardingConfig{
		ExcludedPaths: []string{
			// Stripe Connect routes - users need access to complete onboarding
			"/stripe-connect/",

			// Authentication routes - users need to logout
			"/logout",

			// API routes that don't require onboarding
			"/api/webhook", // Stripe webhooks
			"/api/version", // Version info
			"/version",     // Version info

			// Static assets
			"/assets/",
		},

		ProtectedPaths: []string{
			// Core business functionality that requires payment capability
			"/dashboard",
			"/ebook",
			"/client",
			"/purchase",
			"/file",
			"/settings", // Some settings might be payment-related

			// Transaction and sales management
			"/transactions",
		},
	}
}
