package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	common_types "github.com/kaje94/slek-link/common/pkg/types"
	"github.com/labstack/echo/v4"
)

func GetUserFromCtx(c echo.Context) (userInfo common_types.UserInfo) {
	session, ok := c.Get("session").(*sessions.Session)
	if !ok {
		return common_types.UserInfo{}
	}

	if userInfoJSON, ok := session.Values["userInfo"].(string); ok {
		_ = json.Unmarshal([]byte(userInfoJSON), &userInfo)
	}
	return
}

func GetUserFromCtxWithRedirect(c echo.Context) (userInfo common_types.UserInfo, err error) {
	session, ok := c.Get("session").(*sessions.Session)
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
