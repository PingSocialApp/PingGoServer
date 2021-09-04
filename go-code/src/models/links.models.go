package models

type Link struct {
	Me          *UserBasic `json:"id" db:"me"`
	UserRec     *UserBasic `json:"userRec" db:"user_rec"`
	Permissions int64      `json:"permissions" db:"permissions" binding:"required,min=0,max=8191"`
}

type Request struct {
	ID          string     `json:"id" db:"id"`
	UserRec     *UserBasic `json:"userRec" db:"user_rec"`
	Permissions int64      `json:"permissions" db:"permissions" binding:"required,min=0,max=8191"`
	Me          *UserBasic `json:"me" db:"me"`
}

type OpenRequests struct {
	User   *UserBasic `json:"user"`
	LinkId string     `json:"linkId" db:"link_id"`
}

type LastCheckInLocation struct {
	User      *UserBasic `json:"user"`
	EventName string     `json:"eventName" binding:"ascii,max=50,min=1"`
	EventID   string     `json:"eventId"`
	EventType string     `json:"eventType" binding:"oneof=hangout professional party"`
}
