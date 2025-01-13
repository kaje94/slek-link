package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/kaje94/slek-link/internal/pages"
	"github.com/kaje94/slek-link/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	datastar "github.com/starfederation/datastar/sdk/go"
)

var validate *validator.Validate

func CreateLinkAPIHandler(c echo.Context) error {
	userInfo, err := utils.GetUserFromCtxWithRedirect(c)
	if err != nil {
		return err
	}

	var reqBody struct {
		Name        string `json:"name"`
		ShortCode   string `json:"shortCode"`
		URL         string `json:"url"`
		Description string `json:"description"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	sse.MarshalAndMergeSignals(map[string]any{"linkModalError": ""})

	newLink := models.Link{
		ID:          ulid.Make().String(),
		Name:        reqBody.Name,
		ShortCode:   reqBody.ShortCode,
		LongURL:     reqBody.URL,
		UserID:      &userInfo.ID,
		Description: reqBody.Description,
	}

	if err = validate.Struct(newLink); err != nil {
		sse.MarshalAndMergeSignals(map[string]any{"linkModalError": FormatValidationErrors(err, newLink)})
		return c.String(http.StatusBadRequest, "bad request")
	}

	if err = utils.CreateLink(c, newLink); err == nil {
		sse.Redirect(fmt.Sprintf("/dashboard/%s", newLink.ID))
	} else if strings.Contains(err.Error(), "UNIQUE constraint failed: links.short_code") {
		sse.MarshalAndMergeSignals(map[string]any{"linkModalError": "Short Code already exists. Try a different one."})
		return c.String(http.StatusBadRequest, "Short code already exists")
	} else {
		sse.MarshalAndMergeSignals(map[string]any{"linkModalError": err.Error()})
		return c.String(http.StatusInternalServerError, "Failed to Create Link")
	}

	return nil
}

func DeleteLinkAPIHandler(c echo.Context) error {
	userInfo, err := utils.GetUserFromCtxWithRedirect(c)
	if err != nil {
		return err
	}

	var reqBody struct {
		LinkId string `json:"deleteLinkId"`
	}

	links, err := utils.GetLinksOfUser(c, userInfo.ID)
	if err != nil {
		return err
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if err := utils.DeleteLinkOfUser(c, reqBody.LinkId, userInfo.ID); err == nil {
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
		sse.MarshalAndMergeSignals(map[string]any{"deleteLinkId": ""})
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
