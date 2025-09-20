package main

import (
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/mail"
	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"

	handler "github.com/anglesson/simple-web-server/internal/handler"
	authHandler "github.com/anglesson/simple-web-server/internal/handler/auth"
	"github.com/anglesson/simple-web-server/internal/repository/gorm"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/gov"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

func main() {
	// ========== Infrastructure Initialization ==========
	config.LoadConfigs()
	database.Connect()

	// --- Session Initialization ---
	authKeyHex := os.Getenv("SESSION_AUTH_KEY")
	encKeyHex := os.Getenv("SESSION_ENC_KEY")

	var authKey, encKey []byte
	var err error

	if authKeyHex == "" || encKeyHex == "" {
		log.Println("WARNING: SESSION_AUTH_KEY or SESSION_ENC_KEY not set. Using random keys for this session.")
		authKey = securecookie.GenerateRandomKey(64) // HMAC-SHA-256
		encKey = securecookie.GenerateRandomKey(32)  // AES-256
	} else {
		authKey, err = hex.DecodeString(authKeyHex)
		if err != nil {
			log.Fatalf("FATAL: Failed to decode SESSION_AUTH_KEY: %v. It must be a valid hex-encoded string.", err)
		}
		encKey, err = hex.DecodeString(encKeyHex)
		if err != nil {
			log.Fatalf("FATAL: Failed to decode SESSION_ENC_KEY: %v. It must be a valid hex-encoded string.", err)
		}
	}

	// Validate key lengths to provide helpful warnings.
	if len(authKey) != 64 && len(authKey) != 32 {
		log.Printf("WARNING: SESSION_AUTH_KEY has an invalid length of %d bytes. Recommended: 32 or 64.", len(authKey))
	}
	if len(encKey) != 32 && len(encKey) != 24 && len(encKey) != 16 {
		log.Printf("WARNING: SESSION_ENC_KEY has an invalid length of %d bytes for AES. Valid: 16, 24, or 32.", len(encKey))
	}
	store := sessions.NewCookieStore(authKey, encKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   config.AppConfig.AppMode == "production",
		SameSite: http.SameSiteLaxMode,
	}
	// Create the unified SessionService
	sessionService := service.NewSessionService(store, config.AppConfig.AppName)

	// Template renderer
	templateRenderer := template.DefaultTemplateRenderer()

	// Utils
	encrypter := utils.NewEncrypter()

	// Repositories
	creatorRepository := gorm.NewCreatorRepository(database.DB)
	clientRepository := gorm.NewClientGormRepository()
	userRepository := repository.NewGormUserRepository(database.DB)
	fileRepository := repository.NewGormFileRepository(database.DB)
	purchaseRepository := repository.NewPurchaseRepository()
	transactionRepository := repository.NewTransactionRepository(database.DB)

	// Variáveis para o Mailer
	var mailPort int
	var mailer mail.Mailer
	var emailService *service.EmailService
	var stripeConnectService service.StripeConnectService
	var purchaseService service.PurchaseService
	var transactionService service.TransactionService

	// Services
	commonRFService := gov.NewHubDevService()
	userService := service.NewUserService(userRepository, encrypter)
	subscriptionRepository := gorm.NewSubscriptionGormRepository()
	subscriptionService := service.NewSubscriptionService(subscriptionRepository, commonRFService)
	stripeService := service.NewStripeService()
	paymentGateway := service.NewStripePaymentGateway(stripeService)
	creatorService := service.NewCreatorService(creatorRepository, commonRFService, userService, subscriptionService, paymentGateway)
	clientService := service.NewClientService(clientRepository, creatorRepository, commonRFService)
	s3Storage := storage.NewS3Storage()
	fileService := service.NewFileService(fileRepository, s3Storage)
	ebookService := service.NewEbookService(s3Storage)

	// Mailer para o EmailService
	mailPort, _ = strconv.Atoi(config.AppConfig.MailPort)
	mailer = mail.NewGoMailer(
		config.AppConfig.MailHost,
		mailPort,
		config.AppConfig.MailUsername,
		config.AppConfig.MailPassword)
	emailService = service.NewEmailService(mailer)
	resendDownloadLinkService := service.NewResendDownloadLinkService(transactionRepository, purchaseRepository, emailService)
	stripeConnectService = service.NewStripeConnectService(creatorService)

	// Serviços adicionais - Purchase e Transaction
	purchaseService = service.NewPurchaseService(purchaseRepository, emailService)

	// Transaction Service
	transactionService = service.NewTransactionService(
		transactionRepository,
		purchaseService,
		creatorService,
		stripeService)

	// Handlers
	authHandler := authHandler.NewAuthHandler(userService, sessionService, emailService, templateRenderer)
	clientHandler := handler.NewClientHandler(clientService, creatorService, sessionService, templateRenderer)
	creatorHandler := handler.NewCreatorHandler(creatorService, stripeConnectService, sessionService, templateRenderer)
	settingsHandler := handler.NewSettingsHandler(sessionService, templateRenderer)
	fileHandler := handler.NewFileHandler(fileService, sessionService, templateRenderer)
	ebookHandler := handler.NewEbookHandler(ebookService, creatorService, fileService, s3Storage, sessionService, templateRenderer)
	salesPageHandler := handler.NewSalesPageHandler(ebookService, creatorService, templateRenderer)
	dashboardHandler := handler.NewDashboardHandler(templateRenderer)
	errorHandler := handler.NewErrorHandler(templateRenderer)
	homeHandler := handler.NewHomeHandler(templateRenderer, errorHandler)
	purchaseHandler := handler.NewPurchaseHandler(templateRenderer)
	checkoutHandler := handler.NewCheckoutHandler(templateRenderer, ebookService, clientService, creatorService, commonRFService, emailService, transactionService, purchaseService)
	versionHandler := handler.NewVersionHandler()
	purchaseSalesHandler := handler.NewPurchaseSalesHandler(templateRenderer, purchaseService, sessionService, creatorService, ebookService, resendDownloadLinkService, transactionService)

	stripeHandler := handler.NewStripeHandler(userRepository, subscriptionService, purchaseRepository, purchaseService, emailService, transactionService)
	stripeConnectHandler := handler.NewStripeConnectHandler(stripeConnectService, creatorService, sessionService, templateRenderer)
	transactionHandler := handler.NewTransactionHandler(transactionService, sessionService, creatorService, resendDownloadLinkService, templateRenderer)

	// Initialize rate limiters
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)
	resetPasswordRateLimiter := middleware.NewRateLimiter(5, time.Minute)
	apiRateLimiter := middleware.NewRateLimiter(100, time.Minute)
	uploadRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	// Start cleanup goroutines
	authRateLimiter.CleanupRateLimiter()
	resetPasswordRateLimiter.CleanupRateLimiter()
	apiRateLimiter.CleanupRateLimiter()
	uploadRateLimiter.CleanupRateLimiter()

	r := chi.NewRouter()

	// Apply security headers to all routes
	r.Use(middleware.SecurityHeaders)

	fs := http.FileServer(http.Dir("web/assets"))
	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	})

	// Password reset routes with specific rate limiting (separate from auth)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthGuard(sessionService))
		r.Use(resetPasswordRateLimiter.RateLimitMiddleware)
		r.Get("/forget-password", authHandler.ForgetPasswordView)
		r.Post("/forget-password", authHandler.ForgetPasswordSubmit)
		r.Get("/reset-password", authHandler.ResetPasswordView)
		r.Post("/reset-password", authHandler.ResetPasswordSubmit)
		r.Get("/password-reset-success", func(w http.ResponseWriter, r *http.Request) {
			templateRenderer.View(w, r, "auth/password-reset-success", nil, "guest")
		})
	})

	// Public routes with auth rate limiting (separate from password reset)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthGuard(sessionService))
		r.Use(authRateLimiter.RateLimitMiddleware)
		r.Get("/login", authHandler.LoginView)
		r.Post("/login", authHandler.LoginSubmit)
		r.Get("/register", creatorHandler.RegisterView)
		r.Post("/register", creatorHandler.RegisterCreatorSSR)
		r.Get("/sales/{slug}", salesPageHandler.SalesPageView)
	})

	// Completely public routes (no middleware)
	r.Get("/purchase/download/{id}", purchaseHandler.PurchaseDownloadHandler)
	r.Get("/checkout/{id}", checkoutHandler.CheckoutView)
	r.Get("/purchase/success", checkoutHandler.PurchaseSuccessView)

	// Version routes
	r.Get("/version", versionHandler.VersionText)
	r.Get("/api/version", versionHandler.VersionInfo)

	// Stripe routes with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(apiRateLimiter.RateLimitMiddleware)
		r.Post("/api/create-checkout-session", stripeHandler.CreateCheckoutSession)
		r.Post("/api/webhook", stripeHandler.HandleStripeWebhook)
		r.Post("/api/watermark", handler.WatermarkHandler)
		r.Post("/api/validate-customer", checkoutHandler.ValidateCustomer)
		r.Post("/api/create-ebook-checkout", checkoutHandler.CreateEbookCheckout)
	})

	// Private routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(sessionService))
		r.Use(middleware.StripeOnboardingMiddleware(creatorService, stripeConnectService, sessionService))
		// r.Use(middleware.TrialMiddleware) // TODO: Re-enable this middleware
		r.Use(middleware.SubscriptionMiddleware(subscriptionService))

		r.Post("/logout", authHandler.LogoutSubmit)
		r.Get("/dashboard", dashboardHandler.DashboardView)
		r.Get("/settings", settingsHandler.SettingsView)

		// Ebook routes
		r.Get("/ebook", ebookHandler.IndexView)
		r.Get("/ebook/create", ebookHandler.CreateView)
		r.Post("/ebook/create", ebookHandler.CreateSubmit)
		r.Get("/ebook/edit/{id}", ebookHandler.UpdateView)
		r.Get("/ebook/view/{id}", ebookHandler.ShowView)
		r.Post("/ebook/update/{id}", ebookHandler.UpdateSubmit)
		r.Get("/ebook/preview/{id}", salesPageHandler.SalesPagePreviewView)
		r.Get("/ebook/sales-page/{slug}", salesPageHandler.SalesPageView)
		r.Get("/ebook/{id}/image", ebookHandler.ServeEbookImage)

		// File routes with upload rate limiting
		r.Group(func(r chi.Router) {
			r.Use(uploadRateLimiter.RateLimitMiddleware)
			r.Get("/file", fileHandler.FileIndexView)
			r.Get("/file/upload", fileHandler.FileUploadView)
			r.Post("/file/upload", fileHandler.FileUploadSubmit)
			r.Post("/file/{id}/update", fileHandler.FileUpdateSubmit)
			r.Post("/file/{id}/delete", fileHandler.FileDeleteSubmit)
		})

		// Client routes
		r.Get("/client", clientHandler.ClientIndexView)
		r.Get("/client/new", clientHandler.CreateView)
		r.Post("/client", clientHandler.ClientCreateSubmit)
		r.Get("/client/update/{id}", clientHandler.UpdateView)
		r.Post("/client/update/{id}", clientHandler.ClientUpdateSubmit)
		r.Post("/client/import", clientHandler.ClientImportSubmit)

		// Purchase routes
		r.Post("/purchase/ebook/{id}", purchaseHandler.PurchaseCreateHandler)
		r.Get("/purchase/sales", purchaseSalesHandler.PurchaseSalesList)
		r.Post("/purchase/sales/block-download", purchaseSalesHandler.BlockDownload)
		r.Post("/purchase/sales/unblock-download", purchaseSalesHandler.UnblockDownload)
		r.Post("/purchase/sales/resend-link", purchaseSalesHandler.ResendDownloadLink)

		// Onboarding Stripe Routes
		r.Get("/stripe-connect/welcome", stripeConnectHandler.OnboardingWelcome)
		r.Get("/stripe-connect/complete", stripeConnectHandler.CompleteOnboarding)
		r.Get("/stripe-connect/status", stripeConnectHandler.OnboardingStatus)
		r.Get("/stripe-connect/onboard", stripeConnectHandler.StartOnboarding)

		// Transaction Routes (apenas detalhes acessíveis via vendas)
		r.Get("/transactions/detail", transactionHandler.TransactionDetail)
		r.Post("/transactions/resend-download-link", transactionHandler.ResendDownloadLink)
	})

	r.Get("/", homeHandler.HomeView)

	// r.NotFound(utils.NotFound)
	// r.MethodNotAllowed(utils.MethodNotAllowed)

	// Start server
	log.Printf("Server starting on %s:%s", config.AppConfig.Host, config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, r))
}
