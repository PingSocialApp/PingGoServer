package models

type Link struct {
	Me          *UserBasic `json:"id" db:"me"`
	UserRec     *UserBasic `json:"userRec" db:"user_rec"`
	Permissions int64      `json:"permissions" db:"permissions"`
}

type Request struct {
	ID          string     `json:"id" db:"id"`
	UserRec     *UserBasic `json:"userRec" db:"user_rec"`
	Permissions int64      `json:"permissions" db:"permissions"`
	Me          *UserBasic `json:"me" db:"me"`
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
