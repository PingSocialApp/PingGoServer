package models

type Location struct {
	Latitude  float64
	Longitude float64
}

type UserBasic struct {
	Bio        string
	Profilepic  string
	UID        string
	Name       string
	Location   Location
	NotifToken string
}

type UserCollection struct {
	Users []UserBasic
}

type Socials struct {
	Instagram string
	Snapchat string
	Facebook string
	LinkedIn string
	ProfessionalEmail string
	PersonalEmail string
	Venmo string
	Website string
	Tiktok string
	Phone int64
	Location bool
}
