package utils

import (
	"strings"
	"time"

	"github.com/kaje94/slek-link/internal/models"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
)

func CreateLink(db *gorm.DB, compat valkeycompat.Cmdable, newLink models.Link) error {
	result := db.Create(&newLink)
	if result.Error != nil {
		return result.Error
	}
	CreateSlugCache(compat, newLink.ShortCode, newLink)
	CreateUserLinkCache(compat, *newLink.UserID, newLink.ID, newLink)
	CreateMonthlyClicksCache(compat, newLink.ID, []models.LinkMonthlyClicks{})
	CreateCountryClicksCache(compat, newLink.ID, []models.LinkCountryClicks{})
	DeleteUserLinksCache(compat, *newLink.UserID)
	DeleteDashboardSearchCache(compat, *newLink.UserID)
	return nil
}

func CreateLinkMonthlyClicks(db *gorm.DB, compat valkeycompat.Cmdable, newMonthlyClicks models.LinkMonthlyClicks) error {
	result := db.Create(&newMonthlyClicks)
	if result.Error == nil {
		DeleteMonthlyClicksCache(compat, newMonthlyClicks.LinkID)
	}
	return result.Error
}

func UpdateLinkMonthlyClicks(compat valkeycompat.Cmdable, db *gorm.DB, monthlyClicks models.LinkMonthlyClicks) error {
	result := db.Save(&monthlyClicks)
	if result.Error != nil {
		return result.Error
	}
	DeleteMonthlyClicksCache(compat, monthlyClicks.LinkID)
	return nil
}

func CreateLinkCountryClicks(compat valkeycompat.Cmdable, db *gorm.DB, newCountryClicks models.LinkCountryClicks) error {
	result := db.Create(&newCountryClicks)
	if result.Error == nil {
		DeleteCountryClicksCache(compat, newCountryClicks.LinkID)
	}
	return result.Error
}

func UpdateLinkCountryClicks(compat valkeycompat.Cmdable, db *gorm.DB, countryClicks models.LinkCountryClicks) error {
	result := db.Save(&countryClicks)
	if result.Error != nil {
		return result.Error
	}
	DeleteCountryClicksCache(compat, countryClicks.LinkID)
	return nil
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

	if results := db.Where(&models.LinkMonthlyClicks{LinkID: linkId}).Limit(12).Order("created_at desc").Find(&monthlyClicks); results.Error != nil {
		return nil, results.Error
	}

	monthlyClicksUpdated := []models.LinkMonthlyClicks{}
	for i := 0; i < 12; i++ {
		pastMonthTime := time.Now().AddDate(0, -i, 0)
		pastMonth := int(pastMonthTime.Month())
		pastMonthYear := pastMonthTime.Year()
		found := false
		for _, item := range monthlyClicks {
			if item.Month == pastMonth && item.Year == pastMonthYear {
				found = true
				monthlyClicksUpdated = append(monthlyClicksUpdated, item)
				break
			}
		}

		if !found {
			monthlyClicksUpdated = append(monthlyClicksUpdated, models.LinkMonthlyClicks{
				LinkID: linkId,
				Year:   pastMonthYear,
				Month:  int(pastMonth),
				Count:  0,
			})
		}
	}

	monthlyClicksTrimmed := []models.LinkMonthlyClicks{}
	for _, item := range monthlyClicksUpdated {
		if len(monthlyClicksTrimmed) > 0 || item.Count > 0 {
			monthlyClicksTrimmed = append(monthlyClicksTrimmed, item)
		}
	}

	CreateMonthlyClicksCache(compat, linkId, monthlyClicksTrimmed)

	return monthlyClicksTrimmed, nil
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

	result := db.Where("user_id = ?", userId).Delete(&models.Link{ID: linkId})
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
	CreateUserLinkCache(compat, *link.UserID, link.ID, link)

	return link, nil
}

func GetCountryClicks(compat valkeycompat.Cmdable, db *gorm.DB, linkId string) ([]models.LinkCountryClicks, error) {
	var countryClicks []models.LinkCountryClicks

	err := GetCountryClicksCache(compat, linkId, &countryClicks)
	if err == nil {
		return countryClicks, nil
	}

	if results := db.Where(&models.LinkCountryClicks{LinkID: linkId}).Limit(10).Order("count desc").Find(&countryClicks); results.Error != nil {
		return nil, results.Error
	}

	CreateCountryClicksCache(compat, linkId, countryClicks)

	return countryClicks, nil
}

func UpdateLink(compat valkeycompat.Cmdable, db *gorm.DB, link models.Link) error {
	result := db.Save(&link)
	if result.Error != nil {
		return result.Error
	}
	CreateUserLinkCache(compat, *link.UserID, link.ID, link)
	CreateSlugCache(compat, link.ShortCode, link)
	DeleteUserLinksCache(compat, *link.UserID)
	DeleteDashboardSearchCache(compat, *link.UserID)
	return nil
}

func GetSearchLinks(compat valkeycompat.Cmdable, db *gorm.DB, userId string, keyword string) ([]models.Link, error) {
	var links []models.Link
	keywordLower := strings.ToLower(keyword)

	err := GetDashboardSearchCache(compat, userId, keywordLower, &links)
	if err == nil {
		return links, nil
	}

	if results := db.Where(&models.Link{UserID: &userId}).Where("lower(name) LIKE ?", "%"+keywordLower+"%").Find(&links); results.Error != nil {
		return nil, results.Error
	}

	CreateDashboardSearchCache(compat, userId, keywordLower, links)

	return links, nil
}
