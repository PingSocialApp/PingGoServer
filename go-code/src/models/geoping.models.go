package models

type shareGeoPing struct{
	ID     []string

}
type createGeoPing struct{
	ID    		string 'json:"ref"'
	sentMess 	string
	timeCreate 	int64
	Location 	Location
	isPrivate 	bool
	timeLimit  	int64
}
type Delete	struct{
	ID 	string
}