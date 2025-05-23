package internal

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/kaje94/slek-link/asyncapi/asyncapi"

	"github.com/ThreeDotsLabs/watermill/message"
	watermillMiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/gorilla/sessions"
	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"github.com/kaje94/slek-link/webapp/internal/config"
	"github.com/kaje94/slek-link/webapp/internal/handlers"
	"github.com/kaje94/slek-link/webapp/internal/pages"
	"github.com/kaje94/slek-link/webapp/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var valkeyClient valkey.Client

func RunServer() error {
	// Create new echo router
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

	// configure sentry as a middleware
	setSentryMiddleware(router)

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

	// Setup AMQP asyncAPI server
	go setupAmqp()

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
	public.GET("/s/:slug", handlers.HandleRedirect)

	dashboard := router.Group("/dashboard")
	dashboard.GET("", pages.HandleDashboardsPage)
	dashboard.GET("/:id", pages.HandleLinkDetailsPage)

	auth := router.Group("")
	auth.GET("/login", handlers.HandleLogin)
	auth.GET("/callback", handlers.HandleAuthCallback)
	auth.GET("/logout", handlers.HandlerLogout)

	api := router.Group("/api/datastar")
	api.POST("/upsert-link", handlers.UpsertLinkAPIHandler)
	api.DELETE("/delete-link", handlers.DeleteLinkAPIHandler)
	api.POST("/dashboard-search", handlers.DashboardSearchAPIHandler)
	api.GET("/lazy/dashboard", handlers.DashboardLazyHandler)
	api.GET("/lazy/link-details/:id", handlers.LinkDetailsLazyHandler)
	api.GET("/lazy/link-details/monthly-clicks/:id", handlers.LinkMonthlyLazyHandler)
	api.GET("/lazy/link-details/country-clicks/:id", handlers.LinkCountryLazyHandler)

	router.GET("/api/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

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
			c.Response().Header().Set("Cache-Control", "public, max-age=604800") // 1 week
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
	db, err = gorm.Open(postgres.Open(config.Config.PostgreSqlDsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if !config.Config.IsProd {
		// Perform auto migrate only if it's not production env
		err = db.AutoMigrate(&gormModels.User{}, &gormModels.Link{}, &gormModels.LinkMonthlyClicks{}, &gormModels.LinkCountryClicks{})
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

func setSentryMiddleware(router *echo.Echo) error {
	if config.Config.SentryDsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: config.Config.SentryDsn,
		}); err != nil {
			return fmt.Errorf("failed to initialize sentry", err)
		}

		router.HTTPErrorHandler = func(err error, c echo.Context) {
			// Get status code
			statusCode := http.StatusInternalServerError
			if he, ok := err.(*echo.HTTPError); ok {
				statusCode = he.Code
			}

			// Capture 500 errors
			if statusCode == http.StatusInternalServerError {
				hub := sentry.CurrentHub().Clone()
				hub.Scope().SetRequest(c.Request())
				hub.CaptureException(err)
			}

			// Call default handler to render response
			router.DefaultHTTPErrorHandler(err, c)
		}
	}
	return nil
}

func setupAmqp() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	compat := valkeycompat.NewAdapter(valkeyClient)
	defer stop()

	router, err := asyncapi.GetRouter()
	if err != nil {
		log.Fatalf("error getting router: %s", err)
	}

	amqpPub, err := asyncapi.GetAMQPPublisher(config.Config.AmqpUrl)
	if err != nil {
		log.Fatalf("error getting amqp publisher: %s", err)
	}

	ch, err := amqpPub.Connection().Channel()
	if err != nil {
		log.Fatalf("error getting amqp channel: %s", err)
	}

	urlVisitedQueueName := "url/visited"
	_, err = ch.QueueDeclare(urlVisitedQueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare '%s' queue: %v", urlVisitedQueueName, err)
	}

	poisonQueueName := "url/visited/error"
	_, err = ch.QueueDeclare(poisonQueueName, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare '%s' queue: %v", poisonQueueName, err)
	}

	poisonQueue, err := watermillMiddleware.PoisonQueue(amqpPub, poisonQueueName)
	if err != nil {
		log.Fatalf("error getting error queue: %s", err)
	}

	router.AddMiddleware(
		poisonQueue,
		watermillMiddleware.Recoverer,
		watermillMiddleware.Timeout(time.Second*30),
	)

	amqpSubscriber, err := asyncapi.GetAMQPSubscriber(config.Config.AmqpUrl)
	if err != nil {
		log.Fatalf("error starting amqp subscribers: %s", err)
	}

	// Add asyncAPI handlers
	router.AddNoPublisherHandler("url_visited_handler", urlVisitedQueueName, amqpSubscriber, func(msg *message.Message) error {
		return handlers.HandleUserUrlVisit(compat, db, msg)
	})

	if err = router.Run(ctx); err != nil {
		log.Fatalf("error running watermill router: %s", err)
	}
}
