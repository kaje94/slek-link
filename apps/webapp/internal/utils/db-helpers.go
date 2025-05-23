package utils

import (
	"strings"

	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
)

func CreateLink(db *gorm.DB, compat valkeycompat.Cmdable, newLink gormModels.Link) error {
	result := db.Create(&newLink)
	if result.Error != nil {
		return result.Error
	}
	CreateSlugCache(compat, newLink.ShortCode, newLink)
	CreateUserLinkCache(compat, *newLink.UserID, newLink.ID, newLink)
	CreateMonthlyClicksCache(compat, newLink.ID, []gormModels.LinkMonthlyClicks{})
	CreateCountryClicksCache(compat, newLink.ID, []gormModels.LinkCountryClicks{})
	DeleteDashboardSearchCache(compat, *newLink.UserID)

	var existingLinksCache []gormModels.Link
	GetUserLinksCache(compat, *newLink.UserID, &existingLinksCache)
	if existingLinksCache != nil {
		existingLinksCache = append(existingLinksCache, newLink)
		CreateUserLinksCache(compat, *newLink.UserID, existingLinksCache)
	} else {
		DeleteUserLinksCache(compat, *newLink.UserID)
	}
	return nil
}

func CreateLinkMonthlyClicks(db *gorm.DB, compat valkeycompat.Cmdable, newMonthlyClicks gormModels.LinkMonthlyClicks) error {
	result := db.Create(&newMonthlyClicks)
	if result.Error == nil {
		DeleteMonthlyClicksCache(compat, newMonthlyClicks.LinkID)
	}
	return result.Error
}

func UpdateLinkMonthlyClicks(compat valkeycompat.Cmdable, db *gorm.DB, monthlyClicks gormModels.LinkMonthlyClicks) error {
	result := db.Save(&monthlyClicks)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func CreateLinkCountryClicks(compat valkeycompat.Cmdable, db *gorm.DB, newCountryClicks gormModels.LinkCountryClicks) error {
	result := db.Create(&newCountryClicks)
	if result.Error == nil {
		DeleteCountryClicksCache(compat, newCountryClicks.LinkID)
	}
	return result.Error
}

func UpdateLinkCountryClicks(compat valkeycompat.Cmdable, db *gorm.DB, countryClicks gormModels.LinkCountryClicks) error {
	result := db.Save(&countryClicks)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetLinksOfUser(compat valkeycompat.Cmdable, db *gorm.DB, userId string) ([]gormModels.Link, error) {
	var links []gormModels.Link

	err := GetUserLinksCache(compat, userId, &links)
	if err == nil {
		return links, nil
	}

	if results := db.Where(&gormModels.Link{UserID: &userId}).Find(&links); results.Error != nil {
		return nil, results.Error
	}

	CreateUserLinksCache(compat, userId, links)

	return links, nil
}

func GetLinksMonthlyClicks(compat valkeycompat.Cmdable, db *gorm.DB, linkId string) ([]gormModels.LinkMonthlyClicks, error) {
	var monthlyClicks []gormModels.LinkMonthlyClicks

	err := GetMonthlyClicksCache(compat, linkId, &monthlyClicks)
	if err == nil {
		return monthlyClicks, nil
	}

	if results := db.Where(&gormModels.LinkMonthlyClicks{LinkID: linkId}).Limit(12).Order("created_at desc").Find(&monthlyClicks); results.Error != nil {
		return nil, results.Error
	}

	CreateMonthlyClicksCache(compat, linkId, monthlyClicks)

	return monthlyClicks, nil
}

func GetLinkOfUser(compat valkeycompat.Cmdable, db *gorm.DB, userId, linkId string) (gormModels.Link, error) {
	var link gormModels.Link

	err := GetUserLinkCache(compat, userId, linkId, &link)
	if err == nil {
		return link, nil
	}

	if results := db.Where(&gormModels.Link{UserID: &userId, ID: linkId}).Find(&link); results.Error != nil {
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

	result := db.Where("user_id = ?", userId).Delete(&gormModels.Link{ID: linkId})
	if result.Error != nil {
		return result.Error
	}

	DeleteUserLinksCache(compat, userId)
	DeleteUserLinkCache(compat, userId, linkId)
	DeleteSlugCache(compat, link.ShortCode)
	DeleteDashboardSearchCache(compat, *link.UserID)
	DeleteCountryClicksCache(compat, link.ID)
	DeleteMonthlyClicksCache(compat, link.ID)
	return nil
}

func GetLinkOfSlug(compat valkeycompat.Cmdable, db *gorm.DB, slug string) (gormModels.Link, error) {
	var link gormModels.Link

	err := GetSlugCache(compat, slug, &link)
	if err == nil {
		return link, nil
	}

	if results := db.Where(&gormModels.Link{ShortCode: slug}).Find(&link); results.Error != nil {
		return link, results.Error
	}

	CreateSlugCache(compat, slug, link)
	CreateUserLinkCache(compat, *link.UserID, link.ID, link)

	return link, nil
}

func GetCountryClicks(compat valkeycompat.Cmdable, db *gorm.DB, linkId string) ([]gormModels.LinkCountryClicks, error) {
	var countryClicks []gormModels.LinkCountryClicks

	err := GetCountryClicksCache(compat, linkId, &countryClicks)
	if err == nil {
		return countryClicks, nil
	}

	if results := db.Where(&gormModels.LinkCountryClicks{LinkID: linkId}).Limit(10).Order("count desc").Find(&countryClicks); results.Error != nil {
		return nil, results.Error
	}

	CreateCountryClicksCache(compat, linkId, countryClicks)

	return countryClicks, nil
}

func UpdateLink(compat valkeycompat.Cmdable, db *gorm.DB, link gormModels.Link) error {
	result := db.Save(&link)
	if result.Error != nil {
		return result.Error
	}
	CreateUserLinkCache(compat, *link.UserID, link.ID, link)
	CreateSlugCache(compat, link.ShortCode, link)
	DeleteDashboardSearchCache(compat, *link.UserID)

	var existingLinksCache []gormModels.Link
	GetUserLinksCache(compat, *link.UserID, &existingLinksCache)
	if existingLinksCache != nil {
		updatedLinks := []gormModels.Link{}
		for _, item := range existingLinksCache {
			if item.ID == link.ID {
				updatedLinks = append(updatedLinks, link)
			} else {
				updatedLinks = append(updatedLinks, item)
			}
		}
		CreateUserLinksCache(compat, *link.UserID, updatedLinks)
	} else {
		DeleteUserLinksCache(compat, *link.UserID)
	}
	return nil
}

func GetSearchLinks(compat valkeycompat.Cmdable, db *gorm.DB, userId string, keyword string) ([]gormModels.Link, error) {
	var links []gormModels.Link
	keywordLower := strings.ToLower(keyword)

	err := GetDashboardSearchCache(compat, userId, keywordLower, &links)
	if err == nil {
		return links, nil
	}

	if results := db.Where(&gormModels.Link{UserID: &userId}).Where("lower(name) LIKE ?", "%"+keywordLower+"%").Find(&links); results.Error != nil {
		return nil, results.Error
	}

	CreateDashboardSearchCache(compat, userId, keywordLower, links)

	return links, nil
}
