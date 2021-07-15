package models

type Link struct {
	ID          string `json:"id" db:"id"`
	UserSentUID string `json:"userSentUID" db:"user_sent_uid"`
	UserRecUID  string `json:"userRecUID" db:"user_rec_id"`
	Permissions int64  `json:"int64" db:"permissions"`
	LinkId      string `json:"linkId" db:"link_id"`
}

type Request struct {
	ID    string	`json:"id" db:"id"`
	UserSent string		`json:"userSent" db:"user_sent"`
	UserRec string	`json:"UserRec" db:"user_rec"`
	Permissions int64	`json:"int64" db:"permissions"`
	UID string	`json:"uid" db:"uid"`
}

type LastCheckInLocation struct {
	ID         string `json:"id" db:"id"`
	UserName   string `json:"username" db:"username"`
	Name       string `json:"name" db:"name"`
	UID        string `json:"uid" db:"uid"`
	Bio        string `json:"bio" db:"bio"`
	LinkId     string `json:"linkId" db:"link_id"`
	ProfilePic string `json:"profilePic" db:"profile_pic"`
	EventName  string `json:"eventName" db:"event_name"`
	EventID    string `json:"eventID" db:"event_id"`
	EventType  string `json:"eventType" db:"event_type"`
}