package utils

import (
	"github.com/kaje94/slek-link/internal/models"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
)

func CreateLink(db *gorm.DB, newLink models.Link) error {
	result := db.Create(&newLink)
	return result.Error
}

func CreateLinkMonthlyClicks(db *gorm.DB, newMonthlyClicks models.LinkMonthlyClicks) error {
	result := db.Create(&newMonthlyClicks)

	return result.Error
}

func UpdateLinkMonthlyClicks(compat valkeycompat.Cmdable, db *gorm.DB, newMonthlyClicks models.LinkMonthlyClicks) error {
	result := db.Save(&newMonthlyClicks)
	if result.Error == nil {
		DeleteMonthlyClicksCache(compat, newMonthlyClicks.LinkID)
	}
	return result.Error
}

func GetLinksOfUser(compat valkeycompat.Cmdable, db *gorm.DB, userId string) ([]models.Link, error) {
	var links []models.Link

	err := GetUserLinksCache(compat, userId, &links)
	if err == nil {
		return links, nil
	}

	if results := db.Where(&models.Link{UserID: &userId}).Find(&links); results.Error != nil {
		return nil, results.Error
	}

	CreateUserLinksCache(compat, userId, links)

	return links, nil
}

func GetLinksMonthlyClicks(compat valkeycompat.Cmdable, db *gorm.DB, linkId string) ([]models.LinkMonthlyClicks, error) {
	var monthlyClicks []models.LinkMonthlyClicks

	err := GetMonthlyClicksCache(compat, linkId, &monthlyClicks)
	if err == nil {
		return monthlyClicks, nil
	}

	if results := db.Where(&models.LinkMonthlyClicks{LinkID: linkId}).Limit(12).Order("id desc").Find(&monthlyClicks); results.Error != nil {
		return nil, results.Error
	}

	CreateMonthlyClicksCache(compat, linkId, monthlyClicks)

	return monthlyClicks, nil
}

func GetLinkOfUser(compat valkeycompat.Cmdable, db *gorm.DB, userId, linkId string) (models.Link, error) {
	var link models.Link

	err := GetUserLinkCache(compat, userId, linkId, &link)
	if err == nil {
		return link, nil
	}

	if results := db.Where(&models.Link{UserID: &userId, ID: linkId}).Find(&link); results.Error != nil {
		return link, results.Error
	}

	CreateUserLinkCache(compat, userId, linkId, link)
	CreateSlugCache(compat, link.ShortCode, link)

	return link, nil
}

func DeleteLinkOfUser(compat valkeycompat.Cmdable, db *gorm.DB, linkId, userId string) error {
	link, err := GetLinkOfUser(compat, db, userId, linkId)
	if err != nil {
		return err
	}

	if result := db.Where("user_id = ?", userId).Delete(&models.Link{ID: linkId}); result.Error == nil {
		return result.Error
	}

	DeleteUserLinksCache(compat, userId)
	DeleteUserLinkCache(compat, userId, linkId)
	DeleteSlugCache(compat, link.ShortCode)

	return nil
}

func GetLinkOfSlug(compat valkeycompat.Cmdable, db *gorm.DB, slug string) (models.Link, error) {
	var link models.Link

	err := GetSlugCache(compat, slug, &link)
	if err == nil {
		return link, nil
	}

	if results := db.Where(&models.Link{ShortCode: slug}).Find(&link); results.Error != nil {
		return link, results.Error
	}

	CreateSlugCache(compat, slug, link)

	return link, nil
}
