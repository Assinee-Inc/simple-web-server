package main

import (
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	accounthandler "github.com/Assinee-Inc/docffy/internal/account/handler"
	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountrepo "github.com/Assinee-Inc/docffy/internal/account/repository"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authhandler "github.com/Assinee-Inc/docffy/internal/auth/handler"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/config"
	deliveryhandler "github.com/Assinee-Inc/docffy/internal/delivery/handler"
	deliveryrepo "github.com/Assinee-Inc/docffy/internal/delivery/repository"
	deliverysvc "github.com/Assinee-Inc/docffy/internal/delivery/service"
	libraryhandler "github.com/Assinee-Inc/docffy/internal/library/handler"
	libraryrepo "github.com/Assinee-Inc/docffy/internal/library/repository"
	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	saleshandler "github.com/Assinee-Inc/docffy/internal/sales/handler"
	salesrepo "github.com/Assinee-Inc/docffy/internal/sales/repository"
	salesrepogorm "github.com/Assinee-Inc/docffy/internal/sales/repository/gorm"
	salesvc "github.com/Assinee-Inc/docffy/internal/sales/service"
	sharedhandler "github.com/Assinee-Inc/docffy/internal/shared/handler"
	subscriptionmiddleware "github.com/Assinee-Inc/docffy/internal/subscription/handler/middleware"
	subscriptionrepository "github.com/Assinee-Inc/docffy/internal/subscription/repository"
	subscriptionservice "github.com/Assinee-Inc/docffy/internal/subscription/service"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/Assinee-Inc/docffy/pkg/gov"
	"github.com/Assinee-Inc/docffy/pkg/mail"
	"github.com/Assinee-Inc/docffy/pkg/middleware"
	"github.com/Assinee-Inc/docffy/pkg/storage"
	"github.com/Assinee-Inc/docffy/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func main() {
	// ========== Infrastructure Initialization ==========
	config.LoadConfigs()
	database.Connect()

	// --- Optional Session Initialization (legacy fallback) ---
	authKeyHex := os.Getenv("SESSION_AUTH_KEY")
	encKeyHex := os.Getenv("SESSION_ENC_KEY")

	var authKey, encKey []byte
	var err error

	if authKeyHex == "" || encKeyHex == "" {
		authKey = securecookie.GenerateRandomKey(64)
		encKey = securecookie.GenerateRandomKey(32)
	} else {
		authKey, err = hex.DecodeString(authKeyHex)
		if err != nil {
			log.Fatalf("FATAL: Failed to decode SESSION_AUTH_KEY: %v", err)
		}
		encKey, err = hex.DecodeString(encKeyHex)
		if err != nil {
			log.Fatalf("FATAL: Failed to decode SESSION_ENC_KEY: %v", err)
		}
	}

	store := sessions.NewCookieStore(authKey, encKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   config.AppConfig.AppMode == "production",
		SameSite: http.SameSiteLaxMode,
	}
	sessionService := authsvc.NewSessionService(store, config.AppConfig.AppName)

	// JWT & Google OIDC Auth Services
	tokenService := authsvc.NewTokenService(config.AppConfig.JWTSecret, time.Duration(config.AppConfig.JWTExpirationHours)*time.Hour)
	googleVerifier := authsvc.NewGoogleOIDCVerifier(config.AppConfig.GoogleClientID)

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

	// Services
	commonRFService := gov.NewHubDevService()
	userService := authsvc.NewUserService(userRepository, encrypter)
	subscriptionRepository := subscriptionrepository.NewSubscriptionGormRepository()
	subscriptionService := subscriptionservice.NewSubscriptionService(subscriptionRepository, commonRFService)
	stripeService := subscriptionservice.NewStripeService()
	paymentGateway := subscriptionservice.NewStripePaymentGateway(stripeService)
	creatorService := accountsvc.NewCreatorService(creatorRepository, commonRFService, userService, subscriptionService, paymentGateway)
	clientService := salesvc.NewClientService(clientRepository, creatorRepository)
	s3Storage := storage.NewS3Storage()
	fileService := librarysvc.NewFileService(fileRepository, s3Storage)
	ebookService := librarysvc.NewEbookService(ebookRepository, s3Storage)

	// Mailer
	mailPort, _ := strconv.Atoi(config.AppConfig.MailPort)
	mailer := mail.NewGoMailer(
		config.AppConfig.MailHost,
		mailPort,
		config.AppConfig.MailUsername,
		config.AppConfig.MailPassword,
	)
	salesEmailService := salesvc.NewEmailService(mailer)
	resendDownloadLinkService := salesvc.NewResendDownloadLinkService(transactionRepository, purchaseRepository, salesEmailService)
	downloadService := deliverysvc.NewDownloadService(purchaseRepository, downloadRepository)
	stripeConnectService := accountsvc.NewStripeConnectService(creatorService)

	purchaseService := salesvc.NewPurchaseService(purchaseRepository, salesEmailService)
	transactionService := salesvc.NewTransactionService(
		transactionRepository,
		purchaseService,
		creatorService,
		stripeService,
	)

	// Handlers (100% JSON API)
	authHandler := authhandler.NewAuthHandler(userService, googleVerifier, tokenService, creatorService, subscriptionService, sessionService)
	creatorHandler := accounthandler.NewCreatorHandler(creatorService, stripeConnectService, userService)
	settingsHandler := accounthandler.NewSettingsHandler(creatorService, sessionService)
	dashboardHandler := accounthandler.NewDashboardHandler()
	stripeConnectHandler := accounthandler.NewStripeConnectHandler(stripeConnectService, creatorService, sessionService)
	fileHandler := libraryhandler.NewFileHandler(fileService)
	ebookHandler := libraryhandler.NewEbookHandler(ebookService, creatorService, fileService, s3Storage)
	salesPageHandler := libraryhandler.NewSalesPageHandler(ebookService, creatorService)
	clientHandler := saleshandler.NewClientHandler(clientService, creatorService)
	purchaseHandler := saleshandler.NewPurchaseHandler(ebookService)
	purchaseSalesHandler := saleshandler.NewPurchaseSalesHandler(purchaseService, sessionService, creatorService, ebookService, resendDownloadLinkService, transactionService)
	transactionHandler := saleshandler.NewTransactionHandler(transactionService, sessionService, creatorService, resendDownloadLinkService)
	checkoutHandler := saleshandler.NewCheckoutHandler(ebookService, clientService, clientRepository, creatorService, commonRFService, salesEmailService, transactionService, purchaseService)
	stripeHandler := saleshandler.NewStripeHandler(userRepository, subscriptionService, purchaseRepository, purchaseService, salesEmailService, transactionService, creatorService)
	downloadHandler := deliveryhandler.NewDownloadHandler(downloadService)
	errorHandler := sharedhandler.NewErrorHandler()
	homeHandler := sharedhandler.NewHomeHandler(errorHandler)

	// Rate limiters
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)
	apiRateLimiter := middleware.NewRateLimiter(100, time.Minute)
	authRateLimiter.CleanupRateLimiter()
	apiRateLimiter.CleanupRateLimiter()

	r := chi.NewRouter()
	r.Use(middleware.SecurityHeaders)

	r.NotFound(errorHandler.NotFound)
	r.MethodNotAllowed(errorHandler.MethodNotAllowed)

	fs := http.FileServer(http.Dir("web/assets"))
	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	})

	// Public Auth & Sales Page routes with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(authRateLimiter.RateLimitMiddleware)
		r.Post("/api/auth/google", authHandler.GoogleLogin)
		r.Get("/sales/{id}", salesPageHandler.SalesPageView)
	})

	// Public Checkout & Delivery routes
	r.Get("/purchase/download/{hash_id}", downloadHandler.PurchaseDownloadHandler)
	r.Get("/checkout/{id}", checkoutHandler.CheckoutView)
	r.Get("/purchase/success", checkoutHandler.PurchaseSuccessView)
	r.Get("/", homeHandler.HomeView)

	// Public Stripe & Checkout API routes with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(apiRateLimiter.RateLimitMiddleware)
		r.Post("/api/create-checkout-session", stripeHandler.CreateCheckoutSession)
		r.Post("/api/webhook", stripeHandler.HandleStripeWebhook)
		r.Post("/api/watermark", libraryhandler.WatermarkHandler)
		r.Post("/api/validate-customer", checkoutHandler.ValidateCustomer)
		r.Post("/api/create-ebook-checkout", checkoutHandler.CreateEbookCheckout)
	})

	// Authenticated User Routes (Step 1 Auth completed; Creator Onboarding may still be pending)
	r.Group(func(r chi.Router) {
		r.Use(authmw.AuthMiddleware(tokenService, userRepository, sessionService))

		r.Get("/api/auth/me", authHandler.Me)
		r.Post("/api/auth/logout", authHandler.LogoutSubmit)
		r.Post("/logout", authHandler.LogoutSubmit)
		r.Post("/api/account/onboarding", creatorHandler.OnboardCreator)
	})

	// Authenticated Creator Routes (Requires Creator Onboarding + Stripe Onboarding + Active Subscription)
	r.Group(func(r chi.Router) {
		r.Use(authmw.AuthMiddleware(tokenService, userRepository, sessionService))
		r.Use(accountmw.CreatorOnboardingMiddleware(creatorService))
		r.Use(accountmw.StripeOnboardingMiddleware(creatorService, stripeConnectService, sessionService))
		r.Use(subscriptionmiddleware.SubscriptionMiddleware(subscriptionService))

		r.Get("/dashboard", dashboardHandler.DashboardView)
		r.Get("/settings", settingsHandler.SettingsView)
		r.Post("/settings/social-name", settingsHandler.UpdateSocialName)
		r.Post("/settings/facebook-pixel", settingsHandler.UpdateFacebookPixel)

		// Ebook routes
		r.Get("/ebook", ebookHandler.IndexView)
		r.Get("/ebook/create", ebookHandler.CreateView)
		r.Post("/ebook/create", ebookHandler.CreateSubmit)
		r.Get("/ebook/edit/{id}", ebookHandler.UpdateView)
		r.Get("/ebook/view/{id}", ebookHandler.ShowView)
		r.Post("/ebook/update/{id}", ebookHandler.UpdateSubmit)
		r.Get("/ebook/preview/{id}", salesPageHandler.SalesPagePreviewView)
		r.Get("/ebook/{id}/image", ebookHandler.GetImageURL)
		r.Post("/ebook/delete/{id}", ebookHandler.RemoveEbook)
		r.Post("/ebook/{id}/remove-file/{fileId}", ebookHandler.RemoveFileFromEbook)

		// File routes
		r.Get("/file", fileHandler.FileIndexView)
		r.Post("/file/upload", fileHandler.FileUploadSubmit)
		r.Post("/file/{id}/update", fileHandler.FileUpdateSubmit)
		r.Post("/file/{id}/delete", fileHandler.FileDeleteSubmit)

		// Client routes
		r.Get("/client", clientHandler.ClientIndexView)
		r.Post("/client/{id}/update", clientHandler.ClientUpdateSubmit)
		r.Get("/client/export", clientHandler.ClientExportCSV)

		// Purchase routes
		r.Post("/purchase/ebook/{id}", purchaseHandler.PurchaseCreateHandler)
		r.Get("/purchase/sales", purchaseSalesHandler.PurchaseSalesList)
		r.Post("/purchase/sales/block-download", purchaseSalesHandler.BlockDownload)
		r.Post("/purchase/sales/unblock-download", purchaseSalesHandler.UnblockDownload)
		r.Post("/purchase/sales/resend-link", purchaseSalesHandler.ResendDownloadLink)

		// Stripe Connect Onboarding Routes
		r.Get("/stripe-connect/welcome", stripeConnectHandler.OnboardingWelcome)
		r.Get("/stripe-connect/complete", stripeConnectHandler.CompleteOnboarding)
		r.Get("/stripe-connect/status", stripeConnectHandler.OnboardingStatus)
		r.Get("/stripe-connect/onboard", stripeConnectHandler.StartOnboarding)

		// Transaction Routes
		r.Get("/transactions/detail", transactionHandler.TransactionDetail)
		r.Post("/transactions/resend-download-link", transactionHandler.ResendDownloadLink)
	})

	log.Printf("Server starting on %s:%s", config.AppConfig.Host, config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, r))
}
