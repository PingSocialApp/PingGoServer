package models

import "time"

type Events struct {
	ID          string     `json:"id" db:"id"`
	Creator     *UserBasic `json:"creator,omitempty" db:"creator"`
	EventName   string     `json:"eventName,omitempty" db:"event_name"`
	Description string     `json:"description,omitempty" db:"description"`
	StartTime   time.Time  `json:"startTime,omitempty" db:"start_time"`
	EndTime     time.Time  `json:"endTime,omitempty" db:"end_time"`
	Type        string     `json:"type,omitempty" db:"type"`
	IsPrivate   bool       `json:"isPrivate,omitempty" db:"is_private"`
	Rating      float64    `json:"rate,omitempty" db:"rate"`
	Location    *Location  `json:"location,omitempty" db:"location"`
}

type ShareEvents struct {
	ID      []string `json:"uids" db:"ids"`
	UID     string   `db:"uid"`
	EventID string   `db:"event_id"`
}

type Checkout struct {
	UID     string `db:"uid"`
	EventID string `db:"event_id"`
	Rating  string `json:"rating" db:"rating"`
	Review  string `json:"review" db:"review"`
}
