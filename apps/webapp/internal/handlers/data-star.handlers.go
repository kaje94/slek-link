package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kaje94/slek-link/internal/config"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/kaje94/slek-link/internal/pages"
	"github.com/kaje94/slek-link/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	datastar "github.com/starfederation/datastar/sdk/go"
)

var validate *validator.Validate

func UpsertLinkAPIHandler(c echo.Context) error {
	userInfo, err := utils.GetUserFromCtxWithRedirect(c)
	if err != nil {
		return err
	}

	compat, err := utils.GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	var reqBody struct {
		Name        string `json:"name"`
		ShortCode   string `json:"shortCode"`
		URL         string `json:"url"`
		Description string `json:"description"`
		EditLinkId  string `json:"editLinkId"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	sse.MarshalAndMergeSignals(map[string]any{"linkModalError": ""})

	if reqBody.EditLinkId != "" {
		item, err := utils.GetLinkOfUser(compat, db, userInfo.ID, reqBody.EditLinkId)
		if err != nil {
			return err
		}

		if item.ID == "" {
			return c.String(http.StatusBadRequest, "bad request")
		}
		item.Name = reqBody.Name
		item.ShortCode = reqBody.ShortCode
		item.LongURL = reqBody.URL
		item.Description = reqBody.Description

		if err = validate.Struct(item); err != nil {
			sse.MarshalAndMergeSignals(map[string]any{"linkModalError": FormatValidationErrors(err, item)})
			return c.String(http.StatusBadRequest, "bad request")
		}

		err = utils.UpdateLink(compat, db, item)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: links.short_code") {
				sse.MarshalAndMergeSignals(map[string]any{"linkModalError": "Short Code already exists. Try a different one."})
				return c.String(http.StatusBadRequest, "Short code already exists")
			} else {
				sse.MarshalAndMergeSignals(map[string]any{"linkModalError": err.Error()})
				return c.String(http.StatusInternalServerError, "Failed to Create Link")
			}
		}

		sse.MarshalAndMergeSignals(map[string]any{
			"linkModalOpen": false,
			"name":          "",
			"shortCode":     "",
			"url":           "",
			"description":   "",
			"editLinkId":    "",
		})

		isInDetailsPage := c.QueryParam("isInDetailsPage")
		if isInDetailsPage != "" {
			sse.MergeFragmentTempl(pages.LinkDetailsBodyNameDesc(item), datastar.WithSelectorID(fmt.Sprintf("link-details-body-title-%s", item.ID)))
			sse.MergeFragmentTempl(pages.LinkDetailsBodyLinks(item), datastar.WithSelectorID(fmt.Sprintf("link-details-body-links-%s", item.ID)))
		} else {
			sse.MergeFragmentTempl(pages.DashboardItem(item), datastar.WithSelectorID(fmt.Sprintf("link-item-%s", reqBody.EditLinkId)))
		}
	} else {
		userLinks, err := utils.GetLinksOfUser(compat, db, userInfo.ID)
		if err != nil {
			return err
		}

		if len(userLinks) >= (config.Config.MaxLinksPerUser) {
			sse.MarshalAndMergeSignals(map[string]any{"linkModalError": "Maximum number of links reached"})
			return c.String(http.StatusBadRequest, "maximum number of links reached")
		}

		newLink := models.Link{
			ID:          ulid.Make().String(),
			Name:        reqBody.Name,
			ShortCode:   reqBody.ShortCode,
			LongURL:     reqBody.URL,
			UserID:      &userInfo.ID,
			Description: reqBody.Description,
			Status:      models.ACTIVE,
		}

		if err = validate.Struct(newLink); err != nil {
			sse.MarshalAndMergeSignals(map[string]any{"linkModalError": FormatValidationErrors(err, newLink)})
			return c.String(http.StatusBadRequest, "bad request")
		}

		if err = utils.CreateLink(db, compat, newLink); err == nil {
			sse.Redirect(fmt.Sprintf("/dashboard/%s?isNewLink=true", newLink.ID))
		} else if strings.Contains(err.Error(), "UNIQUE constraint failed: links.short_code") {
			sse.MarshalAndMergeSignals(map[string]any{"linkModalError": "Short Code already exists. Try a different one."})
			return c.String(http.StatusBadRequest, "Short code already exists")
		} else {
			sse.MarshalAndMergeSignals(map[string]any{"linkModalError": err.Error()})
			return c.String(http.StatusInternalServerError, "Failed to Create Link")
		}
	}

	return nil
}

func DeleteLinkAPIHandler(c echo.Context) error {
	userInfo, err := utils.GetUserFromCtxWithRedirect(c)
	if err != nil {
		return err
	}

	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	compat, err := utils.GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	var reqBody struct {
		LinkId string `json:"deleteLinkId"`
	}

	links, err := utils.GetLinksOfUser(compat, db, userInfo.ID)
	if err != nil {
		return err
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if err := utils.DeleteLinkOfUser(compat, db, reqBody.LinkId, userInfo.ID); err == nil {
		redirectToDashboard := c.QueryParam("redirectToDashboard")
		if redirectToDashboard != "" {
			sse.Redirect("/dashboard")
		} else {
			if len(links) > 1 {
				sse.RemoveFragments(fmt.Sprintf("#link-item-%s", reqBody.LinkId))
			} else {
				// Show empty screen if there are no links left
				sse.MergeFragmentTempl(
					pages.DashboardEmpty(),
					datastar.WithSelectorID("dashboard-content-wrap"),
					datastar.WithMergeMode(datastar.FragmentMergeModeInner),
				)
			}
			sse.MarshalAndMergeSignals(map[string]any{"deleteLinkId": "", "searchInput": ""})
		}
	}

	return nil
}

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("url_friendly", func(fl validator.FieldLevel) bool {
		urlFriendlyRegex := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
		return urlFriendlyRegex.MatchString(fl.Field().String())
	})

}

func FormatValidationErrors(err error, obj interface{}) string {
	if err == nil {
		return ""
	}

	var messages []string

	// Check if the error is of type ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		// Use reflection to get custom messages from the struct tags
		objType := reflect.TypeOf(obj)
		for _, fieldError := range validationErrors {
			fieldName := fieldError.StructField()
			// Find the struct field and get the "message" tag
			if field, found := objType.FieldByName(fieldName); found {
				if customMessage := field.Tag.Get("errormgs"); customMessage != "" {
					messages = append(messages, customMessage)
					continue
				}
			}
			// Default error message if "message" tag is not found
			messages = append(messages, fmt.Sprintf("Field '%s' failed validation: %s", fieldError.Field(), fieldError.Tag()))
		}
	}

	// Combine all messages into a single string
	return strings.Join(messages, ", ")
}

func DashboardSearchAPIHandler(c echo.Context) error {
	userInfo, err := utils.GetUserFromCtxWithRedirect(c)
	if err != nil {
		return err
	}

	compat, err := utils.GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	var reqBody struct {
		Search string `json:"searchInput"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if reqBody.Search == "" {
		links, err := utils.GetLinksOfUser(compat, db, userInfo.ID)
		if err != nil {
			return err
		}
		sse.MergeFragmentTempl(pages.DashboardItems(links), datastar.WithSelectorID("dashboard-items"))
	} else {
		links, err := utils.GetSearchLinks(compat, db, userInfo.ID, reqBody.Search)
		if err != nil {
			return err
		}
		if len(links) > 0 {
			sse.MergeFragmentTempl(pages.DashboardItems(links), datastar.WithSelectorID("dashboard-items"))
		} else {
			sse.MergeFragmentTempl(pages.DashboardNoMatchingItems(), datastar.WithSelectorID("dashboard-items"))
		}
	}

	return nil
}
