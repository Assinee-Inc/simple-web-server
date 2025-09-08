package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/mail"
	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/anglesson/simple-web-server/pkg/utils"

	handler "github.com/anglesson/simple-web-server/internal/handler"
	authHandler "github.com/anglesson/simple-web-server/internal/handler/auth"
	"github.com/anglesson/simple-web-server/internal/handler/web"
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

	flashServiceFactory := func(w http.ResponseWriter, r *http.Request) web.FlashMessagePort {
		return web.NewCookieFlashMessage(w, r)
	}

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
	sessionService := service.NewSessionService()
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
	clientHandler := handler.NewClientHandler(clientService, creatorService, flashServiceFactory, templateRenderer)
	creatorHandler := handler.NewCreatorHandler(creatorService, sessionService, templateRenderer)
	settingsHandler := handler.NewSettingsHandler(sessionService, templateRenderer)
	fileHandler := handler.NewFileHandler(fileService, sessionService, templateRenderer, flashServiceFactory)
	ebookHandler := handler.NewEbookHandler(ebookService, creatorService, fileService, s3Storage, flashServiceFactory, templateRenderer)
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
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)         // 10 requests per minute for auth (increased from 5)
	resetPasswordRateLimiter := middleware.NewRateLimiter(5, time.Minute) // 5 requests per minute for password reset (more restrictive for security)
	apiRateLimiter := middleware.NewRateLimiter(100, time.Minute)         // 100 requests per minute for API
	uploadRateLimiter := middleware.NewRateLimiter(10, time.Minute)       // 10 uploads per minute

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
		r.Use(middleware.AuthGuard)
		r.Use(resetPasswordRateLimiter.RateLimitMiddleware) // Separate rate limiting for password reset
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
		r.Use(middleware.AuthGuard)
		r.Use(authRateLimiter.RateLimitMiddleware) // Rate limiting for auth endpoints only
		r.Get("/login", authHandler.LoginView)
		r.Post("/login", authHandler.LoginSubmit)
		r.Get("/register", creatorHandler.RegisterView)
		r.Post("/register", creatorHandler.RegisterCreatorSSR)
		r.Get("/sales/{slug}", salesPageHandler.SalesPageView) // Página de vendas pública
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
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.TrialMiddleware)
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
		r.Get("/ebook/preview/{id}", salesPageHandler.SalesPagePreviewView) // Preview da página de vendas
		r.Get("/ebook/sales-page/{slug}", salesPageHandler.SalesPageView)   // Página de vendas (alias para preview)
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
		r.Get("/stripe-connect/complete", stripeConnectHandler.CompleteOnboarding)
		r.Get("/stripe-connect/status", stripeConnectHandler.OnboardingStatus)
		r.Get("/stripe-connect/onboard", stripeConnectHandler.StartOnboarding)

		// Transaction Routes
		r.Get("/transactions", transactionHandler.TransactionList)
		r.Get("/transactions/detail", transactionHandler.TransactionDetail)
		r.Post("/transactions/resend-download-link", transactionHandler.ResendDownloadLink)
	})

	r.Get("/", homeHandler.HomeView) // Home page deve ser a ultima rota

	r.NotFound(utils.NotFound)
	r.MethodNotAllowed(utils.MethodNotAllowed)

	// Start server
	log.Printf("Server starting on %s:%s", config.AppConfig.Host, config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, r))
}
