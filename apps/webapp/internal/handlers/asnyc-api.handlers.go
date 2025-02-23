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

	var lm asyncapi.UrlVisitedPayload
	err := json.Unmarshal(msg.Payload, &lm)
	if err != nil {
		return err
	}

	monthlyClicks, err := utils.GetLinksMonthlyClicks(compat, db, lm.LinkId)
	if err != nil {
		return err
	}

	var currentMonth models.LinkMonthlyClicks
	for _, item := range monthlyClicks {
		if item.Month == int(time.Now().Month()) && item.Year == time.Now().Year() {
			currentMonth = item
			break
		}
	}

	if currentMonth.ID == 0 {
		// create current month
		month := time.Now().Month()
		monthStr := strconv.Itoa(int(month))
		if month < 10 {
			monthStr = fmt.Sprintf("0%d", month)
		}
		id, err := strconv.Atoi(fmt.Sprintf("%d%s", time.Now().Year(), monthStr))
		if err != nil {
			return err
		}
		currentMonth = models.LinkMonthlyClicks{
			LinkID: lm.LinkId,
			Year:   time.Now().Year(),
			Month:  int(time.Now().Month()),
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

	return nil
}
