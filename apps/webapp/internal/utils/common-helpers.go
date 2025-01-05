package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetDbFromCtx(c echo.Context) (*gorm.DB, error) {
	db, ok := c.Request().Context().Value(DB_CONTEXT_KEY).(*gorm.DB)
	if !ok {
		return nil, c.String(http.StatusInternalServerError, "Database not found in context")
	}
	return db, nil
}

func GetUserFromCtx(c echo.Context) (userInfo models.User) {
	session, ok := c.Get(string(SESSION_CONTEXT_KEY)).(*sessions.Session)
	if !ok {
		return models.User{}
	}

	if userInfoJSON, ok := session.Values["userInfo"].(string); ok {
		_ = json.Unmarshal([]byte(userInfoJSON), &userInfo)
	}
	return
}

func GetUserFromCtxWithRedirect(c echo.Context) (userInfo models.User, err error) {
	session, ok := c.Get(string(SESSION_CONTEXT_KEY)).(*sessions.Session)
	if !ok {
		// TODO: better to redirect towards a more meaningful page
		err = c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	if userInfoJSON, ok := session.Values["userInfo"].(string); ok {
		parseErr := json.Unmarshal([]byte(userInfoJSON), &userInfo)
		if parseErr != nil {
			err = c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
	}
	if !ok || userInfo.ID == "" {
		err = c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	return
}
