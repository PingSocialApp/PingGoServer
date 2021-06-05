package models

type Link struct {
	ID    string
	UserSent UserBasic
	UserRec UserBasic
	Permissions int64
}

type Request struct {
	ID    string
	UserSent UserBasic
	UserRec UserBasic
	Permissions int64
}

type LastCheckInLocation struct {
	UserName string
	UID string
	Profilepic string
	EventName string
	EventID string
	EventType string
}
