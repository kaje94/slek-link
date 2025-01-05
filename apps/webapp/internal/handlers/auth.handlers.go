package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kaje94/slek-link/internal/config"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/kaje94/slek-link/internal/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     config.Config.WebAppConfig.GoogleClientId,
	ClientSecret: config.Config.WebAppConfig.GoogleClientSecret,
	RedirectURL:  fmt.Sprintf("%s/callback", config.Config.WebAppConfig.Url),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

func HandleLogin(c echo.Context) error {
	if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" {
		// if GoogleClientId or GoogleClientSecret is not available, redirect to /callback and login as a test user
		return c.Redirect(http.StatusTemporaryRedirect, "/callback")
	}
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandlerLogout(c echo.Context) error {
	session, ok := c.Get(utils.SESSION_CONTEXT_KEY).(*sessions.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Session not found")
	}
	session.Values["userInfo"] = nil
	session.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func HandleAuthCallback(c echo.Context) error {
	if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" {
		// if GoogleClientId or GoogleClientSecret is not available, redirect to /callback and login as a test user
		userInfo := models.User{ID: "test-user", Name: "Test User", Email: "test-user@email.com"}

		err := saveUser(c, userInfo)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to handle user information")
		}

		session, ok := c.Get(utils.SESSION_CONTEXT_KEY).(*sessions.Session)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Session not found")
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		session.Values["userInfo"] = string(userInfoJSON)
		session.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	state := c.QueryParam("state")
	if state != "state-token" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state")
	}

	code := c.QueryParam("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code, oauth2.ApprovalForce)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to exchange token")
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user info")
	}
	defer resp.Body.Close()

	userInfo := models.User{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse user info")
	}

	session, ok := c.Get(utils.SESSION_CONTEXT_KEY).(*sessions.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Session not found")
	}

	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	err = saveUser(c, userInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to handle user information")
	}

	session.Values["userInfo"] = string(userInfoJSON)
	session.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

// Save user details to database
func saveUser(c echo.Context, userInfo models.User) error {
	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	res := db.FirstOrCreate(userInfo, userInfo)
	return res.Error
}
