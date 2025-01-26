package handlers

import (
	"net/http"

	"github.com/kaje94/slek-link/internal/utils"
	"github.com/labstack/echo/v4"
)

func HandleRedirect(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/404")
	}

	link, err := utils.GetLinkOfSlug(c, slug)
	if err != nil || link.ID == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/404")
	}

	return c.Redirect(http.StatusTemporaryRedirect, link.LongURL)
}
