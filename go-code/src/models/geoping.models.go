package models

type ShareGeoPing struct {
	ID     []string `json:"ids",db:"ids"`
	UID    string
	PingId string `json:"pingId",db:"ping_Id"`
}

type CreateGeoPing struct{
	sentMess 	string	`json:"sentMess",db:"sent_Message"`
	Location 	Location`json:"location"`
	isPrivate 	bool	`json:"isPrivate",db:"is_Private"`
	timeLimit  	int64	`json:"timeLimit",db:"time_Limit"`
}
