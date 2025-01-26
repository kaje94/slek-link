package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kaje94/slek-link/internal/models"
	"github.com/labstack/echo/v4"
)

func saveCache(c echo.Context, cacheKey string, cacheVal any) error {
	valkeyCompat, err := GetValkeyFromCtx(c)
	if err != nil {
		return err
	}
	jsonBytes, err := json.Marshal(cacheVal)
	if err != nil {
		return err
	}

	if valkeyCompat != nil {
		_, err = valkeyCompat.SetNX(context.Background(), cacheKey, jsonBytes, time.Hour).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func getCache(c echo.Context, cacheKey string, data any) error {
	valkeyCompat, err := GetValkeyFromCtx(c)
	if err != nil {
		return err
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

func deleteCache(c echo.Context, cacheKey string) error {
	valkeyCompat, err := GetValkeyFromCtx(c)
	if err != nil {
		return err
	}

	if valkeyCompat != nil {
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

func getCacheKeyForUserLink(userId, linkId string) string {
	return fmt.Sprintf("link-%s-%s", userId, linkId)
}

func getCacheKeyForSlug(slug string) string {
	return fmt.Sprintf("slug-%s", slug)
}

func CreateUserLinksCache(c echo.Context, userId string, links []models.Link) error {
	return saveCache(c, getCacheKeyForUserLinks(userId), links)
}

func GetUserLinksCache(c echo.Context, userId string, links *[]models.Link) error {
	return getCache(c, getCacheKeyForUserLinks(userId), &links)
}

func DeleteUserLinksCache(c echo.Context, userId string) error {
	return deleteCache(c, getCacheKeyForUserLinks(userId))
}

func CreateUserLinkCache(c echo.Context, userId string, linkId string, link models.Link) error {
	return saveCache(c, getCacheKeyForUserLink(userId, linkId), link)
}

func GetUserLinkCache(c echo.Context, userId string, linkId string, links *models.Link) error {
	return getCache(c, getCacheKeyForUserLink(userId, linkId), &links)
}

func DeleteUserLinkCache(c echo.Context, userId string, linkId string) error {
	return deleteCache(c, getCacheKeyForUserLink(userId, linkId))
}

func CreateSlugCache(c echo.Context, slug string, link models.Link) error {
	return saveCache(c, getCacheKeyForSlug(slug), link)
}

func GetSlugCache(c echo.Context, slug string, link *models.Link) error {
	return getCache(c, getCacheKeyForSlug(slug), &link)
}

func DeleteSlugCache(c echo.Context, slug string) error {
	return deleteCache(c, getCacheKeyForSlug(slug))
}
