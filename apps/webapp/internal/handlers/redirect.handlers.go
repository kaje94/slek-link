package handlers

import (
	"context"
	"net/http"
	"slek-link/asyncapi/asyncapi"
	"time"

	"github.com/kaje94/slek-link/internal/config"
	"github.com/kaje94/slek-link/internal/utils"
	"github.com/labstack/echo/v4"
)

func HandleRedirect(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/404")
	}

	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	compat, err := utils.GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	link, err := utils.GetLinkOfSlug(compat, db, slug)
	if err != nil || link.ID == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/404")
	}

	urlVisitedPayload := asyncapi.UrlVisitedPayload{
		LinkId:    link.ID,
		Timestamp: time.Now().Format(time.RFC3339),
		// todo: handle other properties
	}

	amqpPub, err := asyncapi.GetAMQPPublisher(config.Config.AmqpUrl)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = asyncapi.PublishUrlVisited(ctx, amqpPub, urlVisitedPayload)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusTemporaryRedirect, link.LongURL)
}
