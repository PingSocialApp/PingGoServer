package models

import "time"

type GetGeoPing struct {
	UID        string    `json:"uid"`
	Location   *Location `json:"location"`
	Radius     string    `json:"radius"`
	Name       string    `json:"name"`
	Bio        string    `json:"bio"`
	ProfilePic string    `json:"profilePic"`
	Properties string    `json:"properties"`
	Creator    string    `json:"creator"`
	Feature    string    `json:"feature"`
	ID         string    `json:"id"`
}
type Properties struct {
	Entity     string   `json:"entity"`
	ID         string   `json:"id"`
	Message    string   `json:"message"`
	IsPrivate  bool     `json:"isPrivate"`
	TimeCreate time.Time    `json:"timeCreate"`
	eventType  string   `json:"eventType"`
	Creator    *Creator `json:"creator"`
	Rating     int64    `json:"rating"`
	Type       string   `json:"type"`
	StartTime  time.Time   `json:"startTime"`
	EndTime    time.Time    `json:"endTime"`
}
type Creator struct{
	Name       string    `json:"name"`
	ProfilePic string    `json:"profilePic"`
	ID         string    `json:"id"`
}
type geometry struct{
	point []float64 `json:"point"`
}