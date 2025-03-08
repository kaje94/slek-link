package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"github.com/kaje94/slek-link/internal/config"
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
	if config.Config.Valkey.Url == "" {
		return nil, nil
	}
	valkeycompatCmd, ok := c.Get(string(VALKEY_CONTEXT_KEY)).(valkeycompat.Cmdable)
	if !ok || valkeycompatCmd == nil {
		return nil, fmt.Errorf("failed to find valkey in context")
	}
	return valkeycompatCmd, nil
}

func GetUserFromCtx(c echo.Context) (userInfo gormModels.User) {
	session, ok := c.Get(string(SESSION_CONTEXT_KEY)).(*sessions.Session)
	if !ok {
		return gormModels.User{}
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

func GetUserFromCtxWithRedirect(c echo.Context) (userInfo gormModels.User, err error) {
	session, ok := c.Get(string(SESSION_CONTEXT_KEY)).(*sessions.Session)
	originalURL := c.Request().URL.String()
	loginWithOriginalUrl := "/login?originalURL=" + url.QueryEscape(originalURL)
	if !ok {
		err = c.Redirect(http.StatusTemporaryRedirect, loginWithOriginalUrl)
		return
	}

	if userInfoJSON, ok := session.Values["userInfo"].(string); ok {
		parseErr := json.Unmarshal([]byte(userInfoJSON), &userInfo)
		if parseErr != nil {
			err = c.Redirect(http.StatusTemporaryRedirect, loginWithOriginalUrl)
			return
		}
	}
	if !ok || userInfo.ID == "" {
		err = c.Redirect(http.StatusTemporaryRedirect, loginWithOriginalUrl)
	}
	return
}
