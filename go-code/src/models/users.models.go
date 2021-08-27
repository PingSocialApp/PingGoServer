package models

type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude" binding:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude" binding:"longitude"`
}

type UserBasic struct {
	Bio        string    `json:"bio,omitempty" db:"bio" binding:"ascii,max=150,min=0"`
	ProfilePic string    `json:"profilepic,omitempty" db:"profile_pic"`
	UID        string    `json:"uid,omitempty" db:"uid"`
	Name       string    `json:"name,omitempty" db:"name" binding:"ascii,max=25,min=1"`
	Location   *Location `json:"location,omitempty" db:"location"`
	NotifToken string    `json:"notifToken,omitempty" db:"token"`
	CheckedIn  string    `json:"checkedIn" db:"checked_in"`
}

type UserCollection struct {
	Users []*UserBasic `json:"userBasic,omitempty" db:"user_basic"`
}

type Socials struct {
	Instagram         string `json:"instagram"`
	Snapchat          string `json:"snapchat"`
	Facebook          string `json:"facebook"`
	LinkedIn          string `json:"linkedIn"`
	ProfessionalEmail string `json:"professionalEmail"`
	PersonalEmail     string `json:"personalEmail"`
	Venmo             string `json:"venmo"`
	Website           string `json:"web"`
	Tiktok            string `json:"tiktok"`
	Phone             string `json:"phone"`
	Twitter           string `json:"twitter"`
}
