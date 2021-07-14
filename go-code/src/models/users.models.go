package models

type Location struct {
	Latitude  float64	`json:"latitude"`
	Longitude float64	`json:"longitude"`
}

type UserBasic struct {
	Bio        string	`json:"bio"`
	ProfilePic  string	`json:"pic"`
	UID        string   `json:"uid"`
	Name       string	`json:"name"`
	Location   *Location	`json:"location"`
	NotifToken string	`json:"notifToken"`
	CheckedIn  bool 	`json:"checked"`
}

type UserCollection struct {
	Users []*UserBasic	`json:"userBasic"`
}

type Socials struct {
	Instagram string	`json:"instagram"`
	Snapchat string		`json:"snapchat"`
	Facebook string		`json:"facebook"`
	LinkedIn string		`json:"linkedIn"`
	ProfessionalEmail string	`json:"professionalEmail"`
	PersonalEmail string		`json:"personalEmail"`
	Venmo string		`json:"venmo"`
	Website string		`json:"web"`
	Tiktok string		`json:"tiktok"`
	Phone string		`json:"phone"`
	Twitter string `json:"twitter"`
	Location   *Location	`json:"location"`
}
