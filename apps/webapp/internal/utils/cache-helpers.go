package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kaje94/slek-link/internal/config"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

func saveCache(valkeyCompat valkeycompat.Cmdable, cacheKey string, cacheVal any) error {
	jsonBytes, err := json.Marshal(cacheVal)
	if err != nil {
		return err
	}

	if config.Config.Valkey.Url != "" && valkeyCompat != nil {
		_, err = valkeyCompat.SetNX(context.Background(), cacheKey, jsonBytes, time.Hour).Result()
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
	return fmt.Sprintf("links-%s", userId)
}

func getCacheKeyForCountryClicks(linkId string) string {
	return fmt.Sprintf("country-clicks-%s", linkId)
}

func getCacheKeyForUserLink(userId, linkId string) string {
	return fmt.Sprintf("link-%s-%s", userId, linkId)
}

func getCacheKeyForSlug(slug string) string {
	return fmt.Sprintf("slug-%s", slug)
}

func getCacheKeyForMonthlyClicks(linkId string) string {
	return fmt.Sprintf("monthly-clicks-%s", linkId)
}

func CreateCountryClicksCache(compat valkeycompat.Cmdable, linkId string, item []models.LinkCountryClicks) error {
	return saveCache(compat, getCacheKeyForCountryClicks(linkId), item)
}

func GetCountryClicksCache(compat valkeycompat.Cmdable, linkId string, item *[]models.LinkCountryClicks) error {
	return getCache(compat, getCacheKeyForCountryClicks(linkId), &item)
}

func CreateUserLinksCache(compat valkeycompat.Cmdable, userId string, links []models.Link) error {
	return saveCache(compat, getCacheKeyForUserLinks(userId), links)
}

func GetUserLinksCache(compat valkeycompat.Cmdable, userId string, links *[]models.Link) error {
	return getCache(compat, getCacheKeyForUserLinks(userId), &links)
}

func CreateMonthlyClicksCache(compat valkeycompat.Cmdable, linkId string, links []models.LinkMonthlyClicks) error {
	return saveCache(compat, getCacheKeyForMonthlyClicks(linkId), links)
}

func GetMonthlyClicksCache(compat valkeycompat.Cmdable, linkId string, monthlyClicks *[]models.LinkMonthlyClicks) error {
	return getCache(compat, getCacheKeyForMonthlyClicks(linkId), &monthlyClicks)
}

func DeleteMonthlyClicksCache(compat valkeycompat.Cmdable, linkId string) error {
	return deleteCache(compat, getCacheKeyForMonthlyClicks(linkId))
}

func DeleteUserLinksCache(compat valkeycompat.Cmdable, userId string) error {
	return deleteCache(compat, getCacheKeyForUserLinks(userId))
}

func CreateUserLinkCache(compat valkeycompat.Cmdable, userId string, linkId string, link models.Link) error {
	return saveCache(compat, getCacheKeyForUserLink(userId, linkId), link)
}

func GetUserLinkCache(compat valkeycompat.Cmdable, userId string, linkId string, links *models.Link) error {
	return getCache(compat, getCacheKeyForUserLink(userId, linkId), &links)
}

func DeleteUserLinkCache(compat valkeycompat.Cmdable, userId string, linkId string) error {
	return deleteCache(compat, getCacheKeyForUserLink(userId, linkId))
}

func CreateSlugCache(compat valkeycompat.Cmdable, slug string, link models.Link) error {
	return saveCache(compat, getCacheKeyForSlug(slug), link)
}

func GetSlugCache(compat valkeycompat.Cmdable, slug string, link *models.Link) error {
	return getCache(compat, getCacheKeyForSlug(slug), &link)
}

func DeleteSlugCache(compat valkeycompat.Cmdable, slug string) error {
	return deleteCache(compat, getCacheKeyForSlug(slug))
}

func DeleteCountryClicksCache(compat valkeycompat.Cmdable, linkId string) error {
	return deleteCache(compat, getCacheKeyForCountryClicks(linkId))
}
