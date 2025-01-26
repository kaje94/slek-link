package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
)

func GetDbFromCtx(c echo.Context) (*gorm.DB, error) {
	db, ok := c.Request().Context().Value(DB_CONTEXT_KEY).(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("failed to find database in context")
	}
	return db, nil
}

func GetValkeyFromCtx(c echo.Context) (valkeycompat.Cmdable, error) {
	valkeycompatCmd, ok := c.Get(string(VALKEY_CONTEXT_KEY)).(valkeycompat.Cmdable)
	if !ok || valkeycompatCmd == nil {
		return nil, fmt.Errorf("failed to find valkey in context")
	}
	return valkeycompatCmd, nil

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

func GetCSRFTokenFromCtx(c echo.Context) string {
	return c.Get("csrf").(string)
}

func GetPathFromCtx(c echo.Context) string {
	return c.Request().URL.Path
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
