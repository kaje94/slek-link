package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"github.com/kaje94/slek-link/webapp/internal/config"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

var (
	cacheVersion = "v5"
)

func saveCache(valkeyCompat valkeycompat.Cmdable, cacheKey string, cacheVal any) error {
	jsonBytes, err := json.Marshal(cacheVal)
	if err != nil {
		return err
	}

	if config.Config.Valkey.Url != "" && valkeyCompat != nil {
		_, err = valkeyCompat.Set(context.Background(), cacheKey, jsonBytes, time.Minute*10).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func getCache(valkeyCompat valkeycompat.Cmdable, cacheKey string, data any) error {
	if config.Config.Valkey.Url == "" {
		return fmt.Errorf("valkey not configured")
	}
	if valkeyCompat != nil {
		res, err := valkeyCompat.Cache(time.Second).Get(context.Background(), cacheKey).Result()
		if err != nil {
			return err
		}

		if res == "" {
			return fmt.Errorf("no cache found")
		}

		err = json.Unmarshal([]byte(res), &data)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteCache(valkeyCompat valkeycompat.Cmdable, cacheKey string) error {
	if config.Config.Valkey.Url != "" && valkeyCompat != nil {
		_, err := valkeyCompat.Del(context.Background(), cacheKey).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func getCacheKeyForUserLinks(userId string) string {
	return fmt.Sprintf("links-%s-%s", cacheVersion, userId)
}

func getCacheKeyForUserLink(userId, linkId string) string {
	return fmt.Sprintf("link-%s-%s-%s", cacheVersion, userId, linkId)
}

func getCacheKeyForCountryClicks(linkId string) string {
	return fmt.Sprintf("country-clicks-%s-%s", cacheVersion, linkId)
}

func getCacheKeyForSlug(slug string) string {
	return fmt.Sprintf("slug-%s-%s", cacheVersion, slug)
}

func getCacheKeyForMonthlyClicks(linkId string) string {
	return fmt.Sprintf("monthly-%s-clicks-%s", cacheVersion, linkId)
}

func getCacheKeyForDashboardSearchPrefix(userId string) string {
	return fmt.Sprintf("search-%s-%s", cacheVersion, userId)
}

func getCacheKeyForDashboardSearchKeys(userId string) string {
	return fmt.Sprintf("search-%s-%s", cacheVersion, userId)
}

func getCacheKeyForDashboardSearch(userId string, keyword string) string {
	return fmt.Sprintf("%s-%s", getCacheKeyForDashboardSearchPrefix(userId), keyword)
}

func CreateCountryClicksCache(compat valkeycompat.Cmdable, linkId string, item []gormModels.LinkCountryClicks) error {
	return saveCache(compat, getCacheKeyForCountryClicks(linkId), item)
}

func GetCountryClicksCache(compat valkeycompat.Cmdable, linkId string, item *[]gormModels.LinkCountryClicks) error {
	return getCache(compat, getCacheKeyForCountryClicks(linkId), &item)
}

func CreateUserLinksCache(compat valkeycompat.Cmdable, userId string, links []gormModels.Link) error {
	return saveCache(compat, getCacheKeyForUserLinks(userId), links)
}

func GetUserLinksCache(compat valkeycompat.Cmdable, userId string, links *[]gormModels.Link) error {
	return getCache(compat, getCacheKeyForUserLinks(userId), &links)
}

func CreateMonthlyClicksCache(compat valkeycompat.Cmdable, linkId string, links []gormModels.LinkMonthlyClicks) error {
	return saveCache(compat, getCacheKeyForMonthlyClicks(linkId), links)
}

func GetMonthlyClicksCache(compat valkeycompat.Cmdable, linkId string, monthlyClicks *[]gormModels.LinkMonthlyClicks) error {
	return getCache(compat, getCacheKeyForMonthlyClicks(linkId), &monthlyClicks)
}

func DeleteMonthlyClicksCache(compat valkeycompat.Cmdable, linkId string) error {
	return deleteCache(compat, getCacheKeyForMonthlyClicks(linkId))
}

func DeleteUserLinksCache(compat valkeycompat.Cmdable, userId string) error {
	return deleteCache(compat, getCacheKeyForUserLinks(userId))
}

func CreateUserLinkCache(compat valkeycompat.Cmdable, userId string, linkId string, link gormModels.Link) error {
	return saveCache(compat, getCacheKeyForUserLink(userId, linkId), link)
}

func GetUserLinkCache(compat valkeycompat.Cmdable, userId string, linkId string, links *gormModels.Link) error {
	return getCache(compat, getCacheKeyForUserLink(userId, linkId), &links)
}

func DeleteUserLinkCache(compat valkeycompat.Cmdable, userId string, linkId string) error {
	return deleteCache(compat, getCacheKeyForUserLink(userId, linkId))
}

func CreateSlugCache(compat valkeycompat.Cmdable, slug string, link gormModels.Link) error {
	return saveCache(compat, getCacheKeyForSlug(slug), link)
}

func GetSlugCache(compat valkeycompat.Cmdable, slug string, link *gormModels.Link) error {
	return getCache(compat, getCacheKeyForSlug(slug), &link)
}

func DeleteSlugCache(compat valkeycompat.Cmdable, slug string) error {
	return deleteCache(compat, getCacheKeyForSlug(slug))
}

func DeleteCountryClicksCache(compat valkeycompat.Cmdable, linkId string) error {
	return deleteCache(compat, getCacheKeyForCountryClicks(linkId))
}

func CreateDashboardSearchCache(compat valkeycompat.Cmdable, userId string, keyword string, items []gormModels.Link) error {
	cacheKeys := []string{}
	getCache(compat, getCacheKeyForDashboardSearchKeys(userId), &cacheKeys)
	exists := false
	for _, cacheKey := range cacheKeys {
		if cacheKey == getCacheKeyForDashboardSearch(userId, keyword) {
			exists = true
			break
		}
	}
	if !exists {
		cacheKeys = append(cacheKeys, getCacheKeyForDashboardSearch(userId, keyword))
		err := saveCache(compat, getCacheKeyForDashboardSearchKeys(userId), cacheKeys)
		if err != nil {
			return err
		}
	}
	return saveCache(compat, getCacheKeyForDashboardSearch(userId, keyword), items)
}

func GetDashboardSearchCache(compat valkeycompat.Cmdable, userId string, keyword string, links *[]gormModels.Link) error {
	return getCache(compat, getCacheKeyForDashboardSearch(userId, keyword), &links)
}

func DeleteDashboardSearchCache(compat valkeycompat.Cmdable, userId string) error {
	cacheKeys := []string{}
	getCache(compat, getCacheKeyForDashboardSearchKeys(userId), &cacheKeys)
	for _, cacheKey := range cacheKeys {
		err := deleteCache(compat, cacheKey)
		if err != nil {
			return err
		}
	}
	return nil
}
