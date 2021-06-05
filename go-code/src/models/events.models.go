package models

type Events struct {
	ID          string
	Name        string
	Description string
	TimeStart   string
	TimeEnd     string
	Type        string
	IsPrivate   bool
	Rating      float64
	CreatorId   string
	CreatorName string
	Location    Location
}
