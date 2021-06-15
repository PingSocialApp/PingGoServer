package models

type getGeoPing struct {
	UID 		string 'json:"ref"'
	Location 	Location
	Name 		string
	Bio 		string
	ProfilePic 	string
	Postion 	Location
}
