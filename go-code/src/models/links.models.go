package models

type Link struct {
	Me          *UserBasic `json:"id" db:"me"`
	UserRec     *UserBasic `json:"userRec" db:"user_rec"`
	Permissions int64      `json:"permissions" db:"permissions" binding:"min=0,max=4097"`
}

type Request struct {
	ID          string     `json:"id" db:"id" binding:"uuid4"`
	UserRec     *UserBasic `json:"userRec" db:"user_rec"`
	Permissions int64      `json:"permissions" db:"permissions" binding:"min=0,max=4097"`
	Me          *UserBasic `json:"me" db:"me"`
}

type OpenRequests struct {
	User   *UserBasic `json:"user"`
	LinkId string     `json:"linkId" db:"link_id" binding:"uuid4"`
}

type LastCheckInLocation struct {
	User      *UserBasic `json:"user"`
	EventName string     `json:"eventName" binding:"ascii,max=50,min=1"`
	EventID   string     `json:"eventId" binding:"uuid4"`
	EventType string     `json:"eventType" binding:"oneof=hangout professional party"`
}
