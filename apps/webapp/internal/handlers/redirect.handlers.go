package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kaje94/slek-link/asyncapi/asyncapi"

	"github.com/kaje94/slek-link/webapp/internal/config"
	"github.com/kaje94/slek-link/webapp/internal/utils"
	"github.com/labstack/echo/v4"
)

func HandleRedirect(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/404")
	}

	fmt.Println("IP of the user clicking the link", c.RealIP())

	clientIP := c.RealIP()

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
		IpAddress: clientIP,
	}

	amqpPub, err := asyncapi.GetAMQPPublisher(config.Config.AmqpUrl)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	purpose := c.Request().Header.Get("Purpose")
	if purpose != "prefetch" {
		err = asyncapi.PublishUrlVisited(ctx, amqpPub, urlVisitedPayload)
		if err != nil {
			return err
		}
	}

	return c.Redirect(http.StatusTemporaryRedirect, link.LongURL)
}
