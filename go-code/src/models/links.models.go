package models

type Link struct {
	ID    string	`json:"id"`
	UserSentUID string	`json:"userSentUID"`
	UserRecUID string		`json:"userRecUID"`
	Permissions int64		`json:"int64"`

}

type Request struct {
	ID    string	`json:"id"`
	UserSent string		`json:"userSent"`
	UserRec string	`json:"UserRec"`
	Permissions int64	`json:"int64"`
	UID string
}

type LastCheckInLocation struct {
	ID         string `json:"id"`
	UserName   string `json:"username"`
	Name       string `json:"name"`
	UID        string `json:"uid"`
	Bio        string `json:"bio"`
	offset     int64  `json:"offset"`
	limit      int64  `json:"limit"`
	lid        string `json:"rid"`
	LinkId     string `json:"linkId"`
	ProfilePic string `json:"profilePic"`
	EventName  string `json:"eventName"`
	EventID    string `json:"eventID"`
	EventType  string `json:"eventType"`
}