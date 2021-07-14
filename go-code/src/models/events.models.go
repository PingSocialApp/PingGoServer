package models

type Events struct {
	ID          string   `json:"id"`
	UID         string   `json:"uid"`
	Name        string   `json:"name"`
	EventName   string   `json:"eventName"`
	Description string   `json:"description"`
	StartTime   string   `json:"startTime"`
	EndTime     string   `json:"endTime"`
	Type        string   `json:"type"`
	IsPrivate   bool     `json:"isPrivate"`
	Rating      float64  `json:"rate"`
	CreatorId   string   `json:"createId"`
	CreatorName string   `json:"createName"`
	Location    *Location `json:"location"`
}
type ShareEvents struct {
	ID     []string `json:"ids",db:"ids"`
	UID    string
	PingId string `json:"pingId",db:"ping_id"`
}
type Attendee struct{
	ID  string `json:"id"`
	Name  string   `json:"name"`
	UserName string		`json:"user"`
	UID string	`json:"uid"`
	ProfilePic string	`json:"profilePic"`
	Bio string	`json:"bio"`
}

