package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-playground/validator/v10"
	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"github.com/kaje94/slek-link/webapp/internal/config"
	"github.com/kaje94/slek-link/webapp/internal/pages"
	"github.com/kaje94/slek-link/webapp/internal/utils"
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
			sse.MergeFragmentTempl(
				pages.LinkDetailsBodyNameDesc(item, item.ID),
				datastar.WithSelectorID("link-details-body-title"),
			)
			sse.MergeFragmentTempl(
				pages.LinkDetailsBodyLinks(item),
				datastar.WithSelectorID("link-details-body-links"),
			)
		} else {
			sse.MergeFragmentTempl(
				pages.DashboardItem(item),
				datastar.WithSelectorID(fmt.Sprintf("link-item-%s", reqBody.EditLinkId)),
			)
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

		newLink := gormModels.Link{
			ID:          ulid.Make().String(),
			Name:        reqBody.Name,
			ShortCode:   reqBody.ShortCode,
			LongURL:     reqBody.URL,
			UserID:      &userInfo.ID,
			Description: reqBody.Description,
			Status:      gormModels.ACTIVE,
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

	// Delete related monthly clicks entries
	monthlyClickIds := []gormModels.LinkMonthlyClicks{}
	monthlyClicks, _ := utils.GetLinksMonthlyClicks(compat, db, reqBody.LinkId)
	for _, item := range monthlyClicks {
		if item.ID != "" {
			monthlyClickIds = append(monthlyClickIds, gormModels.LinkMonthlyClicks{ID: item.ID})
		}
	}
	if len(monthlyClickIds) > 0 {
		db.Delete(&monthlyClickIds)
	}

	// Delete related country clicks entries
	countryClickIds := []gormModels.LinkCountryClicks{}
	countryClicks, _ := utils.GetCountryClicks(compat, db, reqBody.LinkId)
	for _, item := range countryClicks {
		if item.ID != "" {
			countryClickIds = append(countryClickIds, gormModels.LinkCountryClicks{ID: item.ID})
		}
	}
	if len(countryClicks) > 0 {
		db.Delete(&countryClicks)
	}

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
					datastar.WithViewTransitions(),
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
		sse.MergeFragmentTempl(
			pages.DashboardItems(links),
			datastar.WithSelectorID("dashboard-items"),
			datastar.WithViewTransitions(),
		)
	} else {
		links, err := utils.GetSearchLinks(compat, db, userInfo.ID, reqBody.Search)
		if err != nil {
			return err
		}
		if len(links) > 0 {
			sse.MergeFragmentTempl(
				pages.DashboardItems(links),
				datastar.WithSelectorID("dashboard-items"),
				datastar.WithViewTransitions(),
			)
		} else {
			sse.MergeFragmentTempl(
				pages.DashboardNoMatchingItems(),
				datastar.WithSelectorID("dashboard-items"),
				datastar.WithViewTransitions(),
			)
		}
	}

	return nil
}

func DashboardLazyHandler(c echo.Context) error {
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

	links, err := utils.GetLinksOfUser(compat, db, userInfo.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to find links")
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	if len(links) > 0 {
		sse.MergeFragmentTempl(
			pages.DashboardWithContent(c, links),
			datastar.WithSelectorID("dashboard-content-wrap"),
			datastar.WithViewTransitions(),
		)
	} else {
		sse.MergeFragmentTempl(
			pages.DashboardEmpty(),
			datastar.WithSelectorID("dashboard-content-wrap"),
			datastar.WithViewTransitions(),
		)
	}

	return nil
}

func LinkDetailsLazyHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Id is required")
	}

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

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	link, err := utils.GetLinkOfUser(compat, db, userInfo.ID, id)
	if err != nil {
		sse.Redirect("/404")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to find link")
	}

	sse.MergeFragmentTempl(
		pages.LinkDetailsBodyNameDesc(link, id),
		datastar.WithSelectorID("link-details-body-title"),
		datastar.WithViewTransitions(),
	)
	sse.MergeFragmentTempl(
		pages.LinkDetailsActionButtons(link, false),
		datastar.WithSelectorID("link-details-action-buttons"),
		datastar.WithViewTransitions(),
	)
	sse.MergeFragmentTempl(
		pages.LinkDetailsBodyLinks(link),
		datastar.WithSelectorID("link-details-body-links"),
		datastar.WithViewTransitions(),
	)
	timeAgoStr := humanize.Time(link.CreatedAt)
	sse.MergeFragmentTempl(
		pages.StatSection("link-details-stat-created", "Created", timeAgoStr, fmt.Sprintf("Created on %s", link.CreatedAt.Format(utils.LayoutHuman))),
		datastar.WithSelectorID("link-details-stat-created"),
		datastar.WithViewTransitions(),
	)
	sse.MergeFragmentTempl(
		pages.StatSection("link-details-stat-status", "Status", string(link.Status), ""),
		datastar.WithSelectorID("link-details-stat-status"),
		datastar.WithViewTransitions(),
	)

	return nil
}

func LinkMonthlyLazyHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Id is required")
	}

	compat, err := utils.GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	monthlyClicks, err := utils.GetLinksMonthlyClicks(compat, db, id)
	if err != nil {
		return err
	}

	currentMonth, totalClicks := pages.GetTotalAndCurrentMonthClicks(monthlyClicks)
	totalClicksStr := humanize.Comma(int64(totalClicks))
	currentMonthClicksStr := humanize.Comma(int64(currentMonth.Count))
	clicksTrendChart := pages.CreateClicksTrendChart(monthlyClicks)

	sse.MergeFragmentTempl(
		pages.StatSection("link-details-stat-total-clicks", "Total Clicks", totalClicksStr, "Past 12 months"),
		datastar.WithSelectorID("link-details-stat-total-clicks"),
		datastar.WithViewTransitions(),
	)
	sse.MergeFragmentTempl(
		pages.StatSection("link-details-stat-current-month-clicks", "Clicks This Month", currentMonthClicksStr, ""),
		datastar.WithSelectorID("link-details-stat-current-month-clicks"),
		datastar.WithViewTransitions(),
	)
	sse.MergeFragmentTempl(
		pages.StatLineChartSection("link-details-stat-click-trend", id, "Clicks Trend", "Clicks over the past 12 months", clicksTrendChart),
		datastar.WithSelectorID("link-details-stat-click-trend"),
		datastar.WithViewTransitions(),
	)

	return nil
}

func LinkCountryLazyHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Id is required")
	}

	compat, err := utils.GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	db, err := utils.GetDbFromCtx(c)
	if err != nil {
		return err
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	countryClicks, err := utils.GetCountryClicks(compat, db, id)
	if err != nil {
		return err
	}

	countryClicksChart := pages.CreateBarChart(countryClicks)

	sse.MergeFragmentTempl(
		pages.StatBarChartSection("link-details-stat-country-clicks", id, "Clicks by Country", "Top countries where the users who clicked on the link are from", countryClicksChart),
		datastar.WithSelectorID("link-details-stat-country-clicks"),
		datastar.WithViewTransitions(),
	)

	return nil
}
