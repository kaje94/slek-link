package utils

import (
	"github.com/kaje94/slek-link/internal/models"
	"github.com/labstack/echo/v4"
)

func CreateLink(c echo.Context, newLink models.Link) error {
	db, err := GetDbFromCtx(c)
	if err != nil {
		return err
	}
	result := db.Create(&newLink)
	return result.Error
}

func GetLinksOfUser(c echo.Context, userId string) ([]models.Link, error) {
	var links []models.Link

	db, err := GetDbFromCtx(c)
	if err != nil {
		return nil, err
	}

	err = GetUserLinksCache(c, userId, &links)
	if err == nil {
		return links, nil
	}

	if results := db.Where(&models.Link{UserID: &userId}).Find(&links); results.Error != nil {
		return nil, results.Error
	}

	CreateUserLinksCache(c, userId, links)

	return links, nil
}

func GetLinkOfUser(c echo.Context, userId, linkId string) (models.Link, error) {
	var link models.Link

	db, err := GetDbFromCtx(c)
	if err != nil {
		return link, err
	}

	err = GetUserLinkCache(c, userId, linkId, &link)
	if err == nil {
		return link, nil
	}

	if results := db.Where(&models.Link{UserID: &userId, ID: linkId}).Find(&link); results.Error != nil {
		return link, results.Error
	}

	CreateUserLinkCache(c, userId, linkId, link)
	CreateSlugCache(c, link.ShortCode, link)

	return link, nil
}

func DeleteLinkOfUser(c echo.Context, linkId, userId string) error {
	db, err := GetDbFromCtx(c)
	if err != nil {
		return err
	}

	link, err := GetLinkOfUser(c, userId, linkId)
	if err != nil {
		return err
	}

	if result := db.Where("user_id = ?", userId).Delete(&models.Link{ID: linkId}); result.Error == nil {
		return result.Error
	}

	DeleteUserLinksCache(c, userId)
	DeleteUserLinkCache(c, userId, linkId)
	DeleteSlugCache(c, link.ShortCode)

	return nil
}

func GetLinkOfSlug(c echo.Context, slug string) (models.Link, error) {
	var link models.Link

	db, err := GetDbFromCtx(c)
	if err != nil {
		return link, err
	}

	err = GetSlugCache(c, slug, &link)
	if err == nil {
		return link, nil
	}

	if results := db.Where(&models.Link{ShortCode: slug}).Find(&link); results.Error != nil {
		return link, results.Error
	}

	CreateSlugCache(c, slug, link)

	return link, nil
}
