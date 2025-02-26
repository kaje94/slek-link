package handlers

import (
	"context"
	"fmt"
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

	fmt.Println("IP of the user clicking the link", c.RealIP())

	clientIP := "176.31.84.249"
	//"112.134.213.249" // todo: remove hardcoded ip address
	// clientIP := c.RealIP()

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
	err = asyncapi.PublishUrlVisited(ctx, amqpPub, urlVisitedPayload)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusTemporaryRedirect, link.LongURL)
}
