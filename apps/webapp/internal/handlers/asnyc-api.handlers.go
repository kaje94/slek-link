package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"slek-link/asyncapi/asyncapi"
	"strconv"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/kaje94/slek-link/internal/models"
	"github.com/kaje94/slek-link/internal/utils"
	"github.com/valkey-io/valkey-go/valkeycompat"
	"gorm.io/gorm"
)

func HandleUserUrlVisit(compat valkeycompat.Cmdable, db *gorm.DB, msg *message.Message) error {
	log.Printf("received message payload: %s", string(msg.Payload))
	year := time.Now().Year()
	month := time.Now().Month()

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

	var currentMonth models.LinkMonthlyClicks
	for _, item := range monthlyClicks {
		if item.Month == int(month) && item.Year == year {
			currentMonth = item
			break
		}
	}

	if currentMonth.ID == 0 {
		// create current month
		monthStr := strconv.Itoa(int(month))
		if month < 10 {
			monthStr = fmt.Sprintf("0%d", month)
		}
		id, err := strconv.Atoi(fmt.Sprintf("%d%s", year, monthStr))
		if err != nil {
			return err
		}
		currentMonth = models.LinkMonthlyClicks{
			LinkID: lm.LinkId,
			Year:   year,
			Month:  int(month),
			ID:     id,
			Count:  1,
		}
		err = utils.CreateLinkMonthlyClicks(db, currentMonth)
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

	countryCode, countryName := utils.GetCountry(lm.IpAddress)
	println("payload country", countryCode, countryName)

	if countryCode != "" {
		countryClicks, err := utils.GetCountryClicks(compat, db, lm.LinkId)
		if err != nil {
			return err
		}

		var matchingCountry models.LinkCountryClicks
		for _, item := range countryClicks {
			if item.CountryCode == countryCode {
				matchingCountry = item
				break
			}
		}

		if matchingCountry.ID == "" {
			// create new entry
			countryClicks := models.LinkCountryClicks{
				ID:          fmt.Sprintf("%s-%s", lm.LinkId, countryCode),
				LinkID:      lm.LinkId,
				CountryCode: countryCode,
				CountryName: countryName,
				Count:       1,
			}
			err = utils.CreateLinkCountryClicks(db, countryClicks)
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
	}

	return nil
}
