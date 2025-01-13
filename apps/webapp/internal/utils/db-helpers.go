package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kaje94/slek-link/internal/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type DataHandler struct {
	DB  *gorm.DB
	Ctx context.Context
}

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

func CreateLink(c echo.Context, newLink models.Link) error {
	db, err := GetDbFromCtx(c)
	if err != nil {
		return err
	}
	result := db.Create(&newLink)
	return result.Error
}

func GetLinksOfUser(c echo.Context, userId string) ([]models.Link, error) {
	cacheKey := fmt.Sprintf("links-%s", userId)
	var links []models.Link

	db, err := GetDbFromCtx(c)
	if err != nil {
		return nil, err
	}

	err = getCache(c, cacheKey, &links)
	if err == nil {
		return links, nil
	}

	if results := db.Where(&models.Link{UserID: &userId}).Find(&links); results.Error != nil {
		return nil, results.Error
	}

	saveCache(c, cacheKey, links)

	return links, nil
}

func DeleteLinkOfUser(c echo.Context, linkId, userId string) error {
	db, err := GetDbFromCtx(c)
	if err != nil {
		return err
	}

	if result := db.Where("user_id = ?", userId).Delete(&models.Link{ID: linkId}); result.Error == nil {
		return result.Error
	}

	deleteCache(c, fmt.Sprintf("links-%s", userId))

	return nil
}
