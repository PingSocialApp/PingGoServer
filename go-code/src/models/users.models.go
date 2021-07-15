package models

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type UserBasic struct {
	Bio        string    `json:"bio,omitempty" db:"bio"`
	ProfilePic string    `json:"pic,omitempty" db:"profile_pic"`
	UID        string    `json:"uid,omitempty" db:"uid"`
	Name       string    `json:"name,omitempty" db:"name"`
	Location   *Location `json:"location,omitempty" db:"location"`
	NotifToken string    `json:"notifToken,omitempty" db:"token"`
	CheckedIn  string    `json:"checkedIn,omitempty" db:"checked_in"`
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
