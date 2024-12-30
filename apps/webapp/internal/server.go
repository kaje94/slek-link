package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/kaje94/slek-link/common/pkg/config"
	"github.com/kaje94/slek-link/internal/handlers"
	"github.com/kaje94/slek-link/internal/pages"
)

// runServer runs a new HTTP server with the loaded environment variables.
func RunServer() error {
	// Create a new Echo server.
	router := echo.New()

	// Add Echo middlewares.
	router.Use(middleware.Logger())

	// create store and attach to context
	store := sessions.NewCookieStore([]byte(config.Config.WebAppConfig.CookieSecret))
	store.Options = &sessions.Options{HttpOnly: true, Secure: config.Config.IsProd, SameSite: http.SameSiteLaxMode}
	router.Use(SessionMiddleware(store))

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

	// TODO: Remove once data-star testing is over
	sessionStore := sessions.NewSession(store, "temp")                      // TODO: Remove once data-star testing is over
	pages.SetupExamplesTemplCounter(*router.Router(), sessionStore.Store()) // TODO: Remove once data-star testing is over

	// Create a new server instance with options from environment variables.
	// For more information, see https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", config.Config.WebAppConfig.Port),
		Handler:      router, // handle all Echo routes
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Send log message.
	slog.Info("Starting server...", "port", config.Config.WebAppConfig.Port)

	return server.ListenAndServe()
}

func SessionMiddleware(store sessions.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Retrieve the session
			session, err := store.Get(c.Request(), "session")
			if err != nil {
				slog.Error("Failed to retrieve session", "error", err)
			}

			// Attach the session to the context
			c.Set("session", session)

			// Proceed to the next handler
			return next(c)
		}
	}
}
