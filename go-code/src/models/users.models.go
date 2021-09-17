package models

import "time"

type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude" binding:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude" binding:"longitude"`
}

type UserBasic struct {
	Bio        string    `json:"bio" db:"bio" binding:"omitempty,ascii,max=150,min=1"`
	ProfilePic string    `json:"profilepic" db:"profile_pic" binding:"omitempty"`
	UID        string    `json:"uid" db:"uid" binding:"omitempty"`
	Name       string    `json:"name" db:"name" binding:"omitempty,ascii,max=25,min=1"`
	Location   *Location `json:"location" db:"location" binding:"omitempty"`
	NotifToken string    `json:"notifToken" db:"token" binding:"omitempty"`
	CheckedIn  string    `json:"checkedIn" db:"checked_in" binding:"omitempty"`
	LastOnline time.Time `json:"lastOnline" binding:"omitempty"`
}

type UserCollection struct {
	Users []*UserBasic `json:"userBasic" db:"user_basic" binding:"omitempty"`
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
