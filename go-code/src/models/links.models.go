package models

type Link struct {
	UID         string `json:"id" db:"myUID"`
	UserRecUID  string `json:"userRec" db:"user_rec_id"`
	Permissions int64  `json:"permissions" db:"permissions"`
}

type Request struct {
	ID          string `json:"id" db:"id"`
	UserRec     string `json:"userRec" db:"user_rec"`
	Permissions int64  `json:"int64" db:"permissions"`
	UID         string `json:"uid" db:"uid"`
}

type OpenRequests struct {
	User   *UserBasic `json:"user"`
	LinkId string     `json:"linkId" db:"link_id"`
}

type LastCheckInLocation struct {
	User      *UserBasic `json:"user"`
	EventName string     `json:"eventName"`
	EventID   string     `json:"eventId"`
	EventType string     `json:"eventType"`
}
