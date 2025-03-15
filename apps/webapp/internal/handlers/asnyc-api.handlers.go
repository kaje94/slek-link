package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kaje94/slek-link/asyncapi/asyncapi"

	"github.com/ThreeDotsLabs/watermill/message"
	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"github.com/kaje94/slek-link/webapp/internal/utils"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
)

func HandleUserUrlVisit(compat valkeycompat.Cmdable, db *gorm.DB, msg *message.Message) error {
	log.Printf("received message payload: %s", string(msg.Payload))
	year := time.Now().Year()
	month := 2

	var lm asyncapi.UrlVisitedPayload
	err := json.Unmarshal(msg.Payload, &lm)
	if err != nil {
		return err
	}

	if lm.LinkId == "" {
		return nil
	}

	monthlyClicks, err := utils.GetLinksMonthlyClicks(compat, db, lm.LinkId)
	if err != nil {
		return err
	}

	var currentMonth gormModels.LinkMonthlyClicks
	for _, item := range monthlyClicks {
		if item.Month == int(month) && item.Year == year {
			currentMonth = item
			break
		}
	}

	isNewMonth := false

	if currentMonth.ID == "" {
		// create current month
		isNewMonth = true
		monthStr := strconv.Itoa(int(month))
		if month < 10 {
			monthStr = fmt.Sprintf("0%d", month)
		}
		currentMonth = gormModels.LinkMonthlyClicks{
			LinkID: lm.LinkId,
			Year:   year,
			Month:  int(month),
			ID:     fmt.Sprintf("%s-%d%s", lm.LinkId, year, monthStr),
			Count:  1,
		}
		err = utils.CreateLinkMonthlyClicks(db, compat, currentMonth)
		if err != nil {
			return err
		}
	} else {
		// update current month
		currentMonth.Count += 1
		err = utils.UpdateLinkMonthlyClicks(compat, db, currentMonth)
		if err != nil {
			return err
		}
	}

	var newMonthlyClicksList []gormModels.LinkMonthlyClicks
	if isNewMonth {
		newMonthlyClicksList = append(monthlyClicks, currentMonth)
		// newMonthlyClicksList = append(newMonthlyClicksList, currentMonth)
	} else {
		for _, item := range monthlyClicks {
			if item.Month == int(month) && item.Year == year {
				newMonthlyClicksList = append(newMonthlyClicksList, currentMonth)
			} else {
				newMonthlyClicksList = append(newMonthlyClicksList, item)
			}
		}
	}
	utils.CreateMonthlyClicksCache(compat, lm.LinkId, newMonthlyClicksList)

	countryCode, countryName := utils.GetCountry(lm.IpAddress)
	println("payload country", countryCode, countryName)

	if countryCode != "" {
		countryClicks, err := utils.GetCountryClicks(compat, db, lm.LinkId)
		if err != nil {
			return err
		}

		var matchingCountry gormModels.LinkCountryClicks
		for _, item := range countryClicks {
			if item.CountryCode == countryCode {
				matchingCountry = item
				break
			}
		}

		isNewCountry := false
		var countryClickItem gormModels.LinkCountryClicks

		if matchingCountry.ID == "" {
			// create new entry
			isNewCountry = true
			countryClickItem = gormModels.LinkCountryClicks{
				ID:          fmt.Sprintf("%s-%s", lm.LinkId, countryCode),
				LinkID:      lm.LinkId,
				CountryCode: countryCode,
				CountryName: countryName,
				Count:       1,
			}
			err = utils.CreateLinkCountryClicks(compat, db, countryClickItem)
			if err != nil {
				return err
			}
		} else {
			// update existing entry
			matchingCountry.Count += 1
			err = utils.UpdateLinkCountryClicks(compat, db, matchingCountry)
			if err != nil {
				return err
			}
		}

		var newCountryClicksList []gormModels.LinkCountryClicks
		if isNewCountry {
			newCountryClicksList = append(newCountryClicksList, countryClicks...)
			newCountryClicksList = append(newCountryClicksList, countryClickItem)
		} else {
			for _, item := range countryClicks {
				if item.CountryCode == countryCode {
					newCountryClicksList = append(newCountryClicksList, matchingCountry)
				} else {
					newCountryClicksList = append(newCountryClicksList, item)
				}
			}
		}
		utils.CreateCountryClicksCache(compat, lm.LinkId, newCountryClicksList)
	}

	return nil
}
