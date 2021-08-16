package models

import (
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Events struct {
	ID          string     `json:"id" db:"id" binding:"omitempty,uuid4"`
	Creator     *UserBasic `json:"creator,omitempty" db:"creator"`
	EventName   string     `json:"eventName,omitempty" db:"event_name" binding:"ascii,max=50,min=1"`
	Description string     `json:"description,omitempty" db:"description" binding:"ascii,max=280,min=1"`
	StartTime   time.Time  `json:"startTime,omitempty" db:"start_time" binding:"gte"`
	EndTime     time.Time  `json:"endTime,omitempty" db:"end_time"`
	Type        string     `json:"type,omitempty" db:"type" binding:"oneof=hangout professional party"`
	IsPrivate   bool       `json:"isPrivate,omitempty" db:"is_private"`
	Rating      float64    `json:"rate,omitempty" db:"rate" binding:"max=5,min=0"`
	Location    *Location  `json:"location,omitempty" db:"location"`
}

type ShareEvents struct {
	ID      []string `json:"uids" db:"ids" binding:"min=0,max=30"`
	UID     string   `db:"uid"`
	EventID string   `db:"event_id" binding:"uuid4"`
}

type Checkout struct {
	UID     string  `db:"uid"`
	EventID string  `db:"event_id" binding:"uuid4"`
	Rating  float64 `json:"rating" db:"rating" binding:"min=0,max=5"`
	Review  string  `json:"review" db:"review" binding:"min=1,max=280"`
}

var validEndTime validator.StructLevelFunc = func(sl validator.StructLevel) {
	event := sl.Current().Interface().(Events)

	if event.EndTime.Sub(event.StartTime).Hours() > 24 {
		sl.ReportError(event.EndTime, "endTime", "EndTime", "json", "")
	}
}

func InitCustomEventValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(validEndTime, Events{})
	}
}
