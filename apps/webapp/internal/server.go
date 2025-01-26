package internal

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
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

	// Middleware
	router.Pre(middleware.RemoveTrailingSlash())
	if config.Config.IsProd {
		router.Use(middleware.Secure())
	}
	router.Use(middleware.CSRF())
	router.Use(middleware.Logger())
	router.Use(middleware.Gzip())
	router.Use(middleware.Recover())

	// Session middleware
	store := sessions.NewCookieStore([]byte(config.Config.WebAppConfig.CookieSecret))
	store.Options = &sessions.Options{HttpOnly: true, Secure: config.Config.IsProd, SameSite: http.SameSiteLaxMode}
	router.Use(sessionMiddleware(store))

	// Database middleware
	if err := initDb(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	router.Use(setDBMiddleware)

	// Cache middleware
	if config.Config.Valkey.Url != "" {
		if err := initCache(); err != nil {
			slog.Error("failed to initialize cache", "error", err)
		} else {
			router.Use(setCacheMiddleware)
		}
	}

	// Static file serving with Cache-Control headers
	if config.Config.IsProd {
		router.Use(cacheControlMiddleware) // Add custom caching middleware
	}
	router.Static("/", "./static/public")

	// Routes
	public := router.Group("")
	public.GET("/", pages.HandleLandingPage)
	public.GET("/about-us", pages.HandleAboutUsPage)
	public.GET("/contact-us", pages.HandleContactUsPage)
	public.GET("/faqs", pages.HandleFaqsPage)
	public.GET("/privacy-policy", pages.HandlePrivacyPolicyPage)
	public.GET("/terms-and-conditions", pages.HandleTermsConditionsPage)
	public.GET("/404", pages.Handle404Page)

	dashboard := router.Group("/dashboard")
	dashboard.GET("", pages.HandleDashboardsPage)
	dashboard.GET("/:id", pages.HandleLinkDetailsPage)

	auth := router.Group("")
	auth.GET("/login", handlers.HandleLogin)
	auth.GET("/callback", handlers.HandleAuthCallback)
	auth.GET("/logout", handlers.HandlerLogout)

	api := router.Group("/api")
	api.GET("/hello-world", handlers.ShowContentAPIHandler)
	api.POST("/create-link", handlers.CreateLinkAPIHandler)
	api.DELETE("/delete-link", handlers.DeleteLinkAPIHandler)

	// Server configuration
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", config.Config.WebAppConfig.Port),
		Handler:      router,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("Starting server...", "port", config.Config.WebAppConfig.Port)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("failed to start server", "error", err)
		return err
	}
	return nil
}

func cacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Call the next middleware/handler
		if err := next(c); err != nil {
			return err
		}

		// Set Cache-Control header for specific file types
		path := c.Request().URL.Path
		if strings.HasSuffix(path, ".webmanifest") ||
			strings.HasSuffix(path, ".woff2") ||
			strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".png") ||
			strings.HasSuffix(path, ".webp") ||
			strings.HasSuffix(path, ".ico") ||
			strings.HasSuffix(path, ".svg") {
			c.Response().Header().Set("Cache-Control", "public, max-age=86400") // 1 day
		}

		return nil
	}
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
