package models

import (
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Events struct {
	ID          string     `json:"id" db:"id" binding:"omitempty"`
	Creator     *UserBasic `json:"creator" db:"creator" binding:"omitempty"`
	EventName   string     `json:"eventName" db:"event_name" binding:"required,ascii,max=50,min=1"`
	Description string     `json:"description" db:"description" binding:"required,ascii,max=500,min=1"`
	StartTime   time.Time  `json:"startTime,omitempty" db:"start_time" binding:"required"`
	EndTime     time.Time  `json:"endTime" db:"end_time" binding:"required"`
	Type        string     `json:"type,omitempty" db:"type" binding:"required,oneof=hangout professional party"`
	IsPrivate   bool       `json:"isPrivate" db:"is_private"`
	Rating      float64    `json:"rate" db:"rate" binding:"omitempty,max=5,min=0"`
	Location    *Location  `json:"location" db:"location" binding:"required"`
	IsEnded     bool       `json:"isEnded" binding:"omitempty"`
}

type ShareEvents struct {
	ID      []string `json:"uids" db:"ids" binding:"required,min=0,max=30"`
	UID     string   `db:"uid"`
	EventID string   `db:"event_id"`
	IsNew   bool     `json:"isNew"`
}

type Checkout struct {
	UID     string  `db:"uid"`
	EventID string  `db:"event_id"`
	Rating  float64 `json:"rating" db:"rating" binding:"min=0,max=5"`
	Review  string  `json:"review" db:"review" binding:"min=0,max=280"`
}

var validEventTimes validator.StructLevelFunc = func(sl validator.StructLevel) {
	event := sl.Current().Interface().(Events)

	if time.Since(event.StartTime).Minutes() > 5 {
		sl.ReportError(event.StartTime, "startTime", "StartTime", "json", "")
	}

	timeDiff := event.EndTime.Sub(event.StartTime).Hours()

	if timeDiff > 48.5 || timeDiff < 0 {
		sl.ReportError(event.EndTime, "endTime", "EndTime", "json", "")
	}
}

func InitCustomEventValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(validEventTimes, Events{})
	}
}
