package main

import (
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	accounthandler "github.com/anglesson/simple-web-server/internal/account/handler"
	accountmw "github.com/anglesson/simple-web-server/internal/account/handler/middleware"
	accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"
	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	authhandler "github.com/anglesson/simple-web-server/internal/auth/handler"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	deliveryhandler "github.com/anglesson/simple-web-server/internal/delivery/handler"
	deliveryrepo "github.com/anglesson/simple-web-server/internal/delivery/repository"
	deliverysvc "github.com/anglesson/simple-web-server/internal/delivery/service"
	libraryhandler "github.com/anglesson/simple-web-server/internal/library/handler"
	sharedhandler "github.com/anglesson/simple-web-server/internal/shared/handler"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	librarysvc "github.com/anglesson/simple-web-server/internal/library/service"
	saleshandler "github.com/anglesson/simple-web-server/internal/sales/handler"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesrepogorm "github.com/anglesson/simple-web-server/internal/sales/repository/gorm"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/pkg/gov"
	"github.com/anglesson/simple-web-server/pkg/mail"
	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/pkg/middleware"
	subscriptionmiddleware "github.com/anglesson/simple-web-server/internal/subscription/handler/middleware"
	subscriptionrepository "github.com/anglesson/simple-web-server/internal/subscription/repository"
	subscriptionservice "github.com/anglesson/simple-web-server/internal/subscription/service"
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
	sessionService := authsvc.NewSessionService(store, config.AppConfig.AppName)

	// Template renderer
	templateRenderer := template.DefaultTemplateRenderer()

	// Utils
	encrypter := utils.NewEncrypter()

	// Repositories
	creatorRepository := accountrepo.NewGormCreatorRepository(database.DB)
	clientRepository := salesrepogorm.NewClientGormRepository()
	userRepository := authrepo.NewGormUserRepository(database.DB)
	ebookRepository := libraryrepo.NewGormEbookRepository(database.DB)
	fileRepository := libraryrepo.NewGormFileRepository(database.DB)
	purchaseRepository := salesrepo.NewPurchaseRepository()
	transactionRepository := salesrepo.NewTransactionRepository(database.DB)
	downloadRepository := deliveryrepo.NewGormDownloadRepository()

	// Variáveis para o Mailer
	var mailPort int
	var mailer mail.Mailer
	var authEmailService *authsvc.EmailService
	var salesEmailService *salesvc.EmailService
	var stripeConnectService accountsvc.StripeConnectService
	var purchaseService salesvc.PurchaseService
	var transactionService salesvc.TransactionService

	// Services
	commonRFService := gov.NewHubDevService()
	userService := authsvc.NewUserService(userRepository, encrypter)
	subscriptionRepository := subscriptionrepository.NewSubscriptionGormRepository()
	subscriptionService := subscriptionservice.NewSubscriptionService(subscriptionRepository, commonRFService)
	stripeService := subscriptionservice.NewStripeService()
	paymentGateway := subscriptionservice.NewStripePaymentGateway(stripeService)
	creatorService := accountsvc.NewCreatorService(creatorRepository, commonRFService, userService, subscriptionService, paymentGateway)
	clientService := salesvc.NewClientService(clientRepository, creatorRepository, commonRFService)
	s3Storage := storage.NewS3Storage()
	fileService := librarysvc.NewFileService(fileRepository, s3Storage)
	ebookService := librarysvc.NewEbookService(ebookRepository, s3Storage)

	// Mailer para o EmailService
	mailPort, _ = strconv.Atoi(config.AppConfig.MailPort)
	mailer = mail.NewGoMailer(
		config.AppConfig.MailHost,
		mailPort,
		config.AppConfig.MailUsername,
		config.AppConfig.MailPassword)
	authEmailService = authsvc.NewEmailService(mailer)
	salesEmailService = salesvc.NewEmailService(mailer)
	resendDownloadLinkService := salesvc.NewResendDownloadLinkService(transactionRepository, purchaseRepository, salesEmailService)
	downloadService := deliverysvc.NewDownloadService(purchaseRepository, downloadRepository)
	stripeConnectService = accountsvc.NewStripeConnectService(creatorService)

	// Serviços adicionais - Purchase e Transaction
	purchaseService = salesvc.NewPurchaseService(purchaseRepository, salesEmailService)

	// Transaction Service
	transactionService = salesvc.NewTransactionService(
		transactionRepository,
		purchaseService,
		creatorService,
		stripeService)

	// Handlers
	authHandler := authhandler.NewAuthHandler(userService, sessionService, authEmailService, templateRenderer)
	clientHandler := saleshandler.NewClientHandler(clientService, creatorService, sessionService, templateRenderer)
	creatorHandler := accounthandler.NewCreatorHandler(creatorService, stripeConnectService, sessionService, templateRenderer)
	settingsHandler := accounthandler.NewSettingsHandler(sessionService, templateRenderer)
	fileHandler := libraryhandler.NewFileHandler(fileService, sessionService, templateRenderer)
	ebookHandler := libraryhandler.NewEbookHandler(ebookService, creatorService, fileService, s3Storage, sessionService, templateRenderer)
	salesPageHandler := libraryhandler.NewSalesPageHandler(ebookService, creatorService, templateRenderer)
	dashboardHandler := accounthandler.NewDashboardHandler(templateRenderer)
	errorHandler := sharedhandler.NewErrorHandler(templateRenderer)
	homeHandler := sharedhandler.NewHomeHandler(templateRenderer, errorHandler)
	downloadHandler := deliveryhandler.NewDownloadHandler(downloadService, templateRenderer)
	purchaseHandler := saleshandler.NewPurchaseHandler(templateRenderer, ebookService)
	checkoutHandler := saleshandler.NewCheckoutHandler(templateRenderer, ebookService, clientService, creatorService, commonRFService, salesEmailService, transactionService, purchaseService)
	// versionHandler := handler.NewVersionHandler()
	purchaseSalesHandler := saleshandler.NewPurchaseSalesHandler(templateRenderer, purchaseService, sessionService, creatorService, ebookService, resendDownloadLinkService, transactionService)

	stripeHandler := saleshandler.NewStripeHandler(userRepository, subscriptionService, purchaseRepository, purchaseService, salesEmailService, transactionService, creatorService)
	stripeConnectHandler := accounthandler.NewStripeConnectHandler(stripeConnectService, creatorService, sessionService, templateRenderer)
	transactionHandler := saleshandler.NewTransactionHandler(transactionService, sessionService, creatorService, resendDownloadLinkService, templateRenderer)

	// Initialize rate limiters
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)
	resetPasswordRateLimiter := middleware.NewRateLimiter(5, time.Minute)
	apiRateLimiter := middleware.NewRateLimiter(100, time.Minute)
	// uploadRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	// Start cleanup goroutines
	authRateLimiter.CleanupRateLimiter()
	resetPasswordRateLimiter.CleanupRateLimiter()
	apiRateLimiter.CleanupRateLimiter()
	// uploadRateLimiter.CleanupRateLimiter()

	r := chi.NewRouter()

	// Apply security headers to all routes
	r.Use(middleware.SecurityHeaders)

	fs := http.FileServer(http.Dir("web/assets"))
	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	})

	// Password reset routes with specific rate limiting (separate from auth)
	r.Group(func(r chi.Router) {
		r.Use(authmw.AuthGuard(sessionService))
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
		r.Use(authmw.AuthGuard(sessionService))
		r.Use(authRateLimiter.RateLimitMiddleware)
		r.Get("/login", authHandler.LoginView)
		r.Post("/login", authHandler.LoginSubmit)
		r.Get("/register", creatorHandler.RegisterView)
		r.Post("/register", creatorHandler.RegisterCreatorSSR)
		r.Get("/sales/{id}", salesPageHandler.SalesPageView)
	})

	// Completely public routes (no middleware)
	r.Get("/purchase/download/{hash_id}", downloadHandler.PurchaseDownloadHandler)
	r.Get("/checkout/{id}", checkoutHandler.CheckoutView)
	r.Get("/purchase/success", checkoutHandler.PurchaseSuccessView)

	// Version routes
	// r.Get("/version", versionHandler.VersionText)
	// r.Get("/api/version", versionHandler.VersionInfo)

	// Stripe routes with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(apiRateLimiter.RateLimitMiddleware)
		r.Post("/api/create-checkout-session", stripeHandler.CreateCheckoutSession)
		r.Post("/api/webhook", stripeHandler.HandleStripeWebhook)
		r.Post("/api/watermark", libraryhandler.WatermarkHandler)
		r.Post("/api/validate-customer", checkoutHandler.ValidateCustomer)
		r.Post("/api/create-ebook-checkout", checkoutHandler.CreateEbookCheckout)
	})

	// Private routes
	r.Group(func(r chi.Router) {
		r.Use(authmw.AuthMiddleware(sessionService))
		r.Use(accountmw.StripeOnboardingMiddleware(creatorService, stripeConnectService, sessionService))
		// r.Use(subscriptionmiddleware.TrialMiddleware) // TODO: Re-enable this middleware
		r.Use(subscriptionmiddleware.SubscriptionMiddleware(subscriptionService))

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
		r.Get("/ebook/{id}/image", ebookHandler.ServeEbookImage)
		r.Post("/ebook/delete/{id}", ebookHandler.RemoveEbook)
		r.Post("/ebook/{id}/remove-file/{fileId}", ebookHandler.RemoveFileFromEbook)

		// File routes with upload rate limiting
		r.Group(func(r chi.Router) {
			r.Get("/file", fileHandler.FileIndexView)
			r.Get("/file/upload", fileHandler.FileUploadView)
			r.Post("/file/upload", fileHandler.FileUploadSubmit)
			r.Post("/file/{id}/update", fileHandler.FileUpdateSubmit)
			r.Post("/file/{id}/delete", fileHandler.FileDeleteSubmit)
		})

		// Client routes
		r.Get("/client", clientHandler.ClientIndexView)
		r.Post("/client", clientHandler.ClientCreateSubmit)
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
