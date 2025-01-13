package internal

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gorilla/sessions"
	"github.com/kaje94/slek-link/internal/config"
	"github.com/kaje94/slek-link/internal/handlers"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/kaje94/slek-link/internal/pages"
	"github.com/kaje94/slek-link/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var valkeyClient valkey.Client

func RunServer() error {
	router := echo.New()

	// Add Logger middlewares.
	router.Use(middleware.Logger())

	// create store and attach to context
	store := sessions.NewCookieStore([]byte(config.Config.WebAppConfig.CookieSecret))
	store.Options = &sessions.Options{HttpOnly: true, Secure: config.Config.IsProd, SameSite: http.SameSiteLaxMode}
	router.Use(sessionMiddleware(store))

	// init db and attach db middleware
	if err := initDb(); err != nil {
		return err
	}
	router.Use(setDBMiddleware)

	// init cache and attach db middleware if valkey address is provided
	if config.Config.Valkey.Url != "" {
		if err := initCache(); err != nil {
			return err
		}
		router.Use(setCacheMiddleware)
	}

	// Handle static files.
	router.Static("/", "./static/public")

	// handle pages
	router.GET("/", pages.HandleLandingPage)
	router.GET("/about-us", pages.HandleAboutUsPage)
	router.GET("/contact-us", pages.HandleContactUsPage)
	router.GET("/faqs", pages.HandleFaqsPage)
	router.GET("/privacy-policy", pages.HandlePrivacyPolicyPage)
	router.GET("/terms-and-conditions", pages.HandleTermsConditionsPage)
	router.GET("/dashboard", pages.HandleDashboardsPage)
	router.GET("/dashboard/:id", pages.HandleLinkDetailsPage)

	// Auth handlers
	router.GET("/login", handlers.HandleLogin)
	router.GET("/callback", handlers.HandleAuthCallback)
	router.GET("/logout", handlers.HandlerLogout)

	// Handle API endpoints.
	router.GET("/api/hello-world", handlers.ShowContentAPIHandler)

	// Data-star api handlers
	router.POST("/api/create-link", handlers.CreateLinkAPIHandler)
	router.DELETE("/api/delete-link", handlers.DeleteLinkAPIHandler)

	// TODO: Remove once data-star testing is over
	sessionStore := sessions.NewSession(store, "temp")                      // TODO: Remove once data-star testing is over
	pages.SetupExamplesTemplCounter(*router.Router(), sessionStore.Store()) // TODO: Remove once data-star testing is over

	// Create a new server instance with options from environment variables.
	// For more information, see https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", config.Config.WebAppConfig.Port),
		Handler:      router, // handle all Echo routes
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	slog.Info("Starting server...", "port", config.Config.WebAppConfig.Port)

	return server.ListenAndServe()
}

func sessionMiddleware(store sessions.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Retrieve the session
			session, err := store.Get(c.Request(), string(utils.SESSION_CONTEXT_KEY))
			if err != nil {
				slog.Error("Failed to retrieve session", "error", err)
			}
			// Attach the session to the context
			c.Set(string(utils.SESSION_CONTEXT_KEY), session)
			return next(c)
		}
	}
}

func initDb() error {
	var err error
	db, err = gorm.Open(sqlite.Open("local.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if !config.Config.IsProd {
		// Perform auto migrate only if it's not production env
		err = db.AutoMigrate(&models.User{}, &models.Link{})
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		slog.Info("Database migrated successfully.")
	}
	return nil
}

func initCache() error {
	var err error
	valkeyClient, err = valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{config.Config.Valkey.Url},
	})
	if err != nil {
		log.Fatalf("Could not initialize valkey cache: %v", err)
	}

	return nil
}

func setDBMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		timeoutContext, _ := context.WithTimeout(context.Background(), 10*time.Second)
		ctx := context.WithValue(c.Request().Context(), utils.DB_CONTEXT_KEY, db.WithContext(timeoutContext))
		req := c.Request().WithContext(ctx)
		c.SetRequest(req)
		return next(c)
	}
}

func setCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		compat := valkeycompat.NewAdapter(valkeyClient)
		c.Set(string(utils.VALKEY_CONTEXT_KEY), compat)
		return next(c)
	}
}
