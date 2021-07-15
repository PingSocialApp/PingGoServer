package models

import (
	"time"
)

type GeoPingProp struct {
	Creator     *User     `json:"creator"`
	SentMessage string    `json:"sentMessage"`
	IsPrivate   bool      `json:"isPrivate"`
	TimeCreate  time.Time `json:"timeCreate"`
	ID          string    `json:"id"`
}

type GeoJson struct {
	Properties interface{} `json:"properties"`
	Geometry   *Geometry   `json:"geometry"`
}

type EventProp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsPrivate bool      `json:"isPrivate"`
	EventType string    `json:"eventType"`
	Creator   *User     `json:"creator"`
	Rating    int64     `json:"rating"`
	Type      string    `json:"type"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type User struct {
	Name       string `json:"name"`
	ProfilePic string `json:"profilePic"`
	ID         string `json:"id"`
	Bio        string `json:"bio,omitempty"`
}

type Geometry struct {
	Coordinates []float64 `json:"coordinates"`
	Type        string    `json:"type"`
}

func GetNewGeometry(xCord float64, yCord float64) *Geometry {
	return &Geometry{
		Coordinates: []float64{xCord, yCord},
		Type:        "point",
	}
}
