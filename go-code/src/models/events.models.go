package models

import "time"

type Events struct {
	ID          string    `json:"id" db:"id"`
	UID         string    `json:"uid" db:"uid"`
	EventName   string    `json:"eventName" db:"event_name"`
	Description string    `json:"description,omitempty" db:"description"`
	StartTime   time.Time `json:"startTime,omitempty" db:"start_time"`
	EndTime     time.Time `json:"endTime,omitempty" db:"end_time"`
	Type        string    `json:"type" db:"type"`
	IsPrivate   bool      `json:"isPrivate,omitempty" db:"is_private"`
	Rating      float64   `json:"rate,omitempty" db:"rate"`
	CreatorId   string    `json:"createId,omitempty" db:"uid"`
	CreatorName string    `json:"createName,omitempty" db:"createName"`
	Location    *Location `json:"location,omitempty" db:"location"`
}

type ShareEvents struct {
	ID      []string `db:"ids"`
	UID     string   `db: "uid"`
	EventID string   `db:"event_id"`
}

type Attendee struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	UID        string `json:"uid"`
	ProfilePic string `json:"profilePic"`
	Bio        string `json:"bio"`
}
