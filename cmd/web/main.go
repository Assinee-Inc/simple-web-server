package main

import (
	"log"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/handlers/web"
	"github.com/anglesson/simple-web-server/internal/repositories/gorm"
	"github.com/anglesson/simple-web-server/internal/services"
	"github.com/anglesson/simple-web-server/pkg/gov"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/handlers"
	"github.com/anglesson/simple-web-server/internal/handlers/middleware"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/go-chi/chi/v5"
)

func main() {
	// ========== Infrastructure Initialization ==========
	config.LoadConfigs()
	database.Connect()

	flashServiceFactory := func(w http.ResponseWriter, r *http.Request) web.FlashMessagePort {
		return web.NewCookieFlashMessage(w, r)
	}

	// Repositories
	creatorRepository := gorm.NewCreatorRepository()
	clientRepository := gorm.NewClientGormRepository()

	// ========== Application Initialization ==========
	commonRFService := gov.NewHubDevService()
	clientService := services.NewClientService(clientRepository, creatorRepository, commonRFService)
	clientHandler := handlers.NewClientHandler(clientService, flashServiceFactory)

	r := chi.NewRouter()

	fs := http.FileServer(http.Dir("web/assets"))
	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	})

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthGuard)
		r.Get("/login", handlers.LoginView)
		r.Post("/login", handlers.LoginSubmit)
		r.Get("/register", handlers.RegisterView)
		r.Post("/register", handlers.RegisterSubmit)
		r.Get("/forget-password", handlers.ForgetPasswordView)
		r.Post("/forget-password", handlers.ForgetPasswordSubmit)
		r.Get("/purchase/download/{id}", handlers.PurchaseDownloadHandler)
	})

	// Stripe routes
	r.Post("/api/create-checkout-session", handlers.CreateCheckoutSession)
	r.Post("/api/webhook", handlers.HandleStripeWebhook)

	// Private routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.TrialMiddleware)

		r.Post("/logout", handlers.LogoutSubmit)
		r.Get("/dashboard", handlers.DashboardView)
		r.Get("/settings", handlers.SettingsView)

		// Ebook routes
		r.Get("/ebook", handlers.EbookIndexView)
		r.Get("/ebook/create", handlers.EbookCreateView)
		r.Post("/ebook/create", handlers.EbookCreateSubmit)
		r.Get("/ebook/edit/{id}", handlers.EbookUpdateView)
		r.Get("/ebook/view/{id}", handlers.EbookShowView)
		r.Post("/ebook/update/{id}", handlers.EbookUpdateSubmit)

		// Client routes
		r.Get("/client", clientHandler.ClientIndexView)
		r.Get("/client/new", clientHandler.CreateView)
		r.Post("/client", clientHandler.ClientCreateSubmit)
		r.Get("/client/update/{id}", clientHandler.UpdateView)
		r.Post("/client/update/{id}", clientHandler.ClientUpdateSubmit)
		r.Post("/client/import", clientHandler.ClientImportSubmit)

		// Purchase routes
		r.Post("/purchase/ebook/{id}", handlers.PurchaseCreateHandler)
		r.Get("/purchase/download/{id}", handlers.PurchaseDownloadHandler)
		r.Get("/send", handlers.SendViewHandler)
	})

	r.Get("/", handlers.HomeView) // Home page deve ser a ultima rota

	port := config.AppConfig.Port

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Println("Starting server on :" + port)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
