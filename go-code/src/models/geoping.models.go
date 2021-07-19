package models

type ShareGeoPing struct {
	ID      []string   `json:"ids" db:"ids"`
	Creator *UserBasic `db:"creator"`
	PingID  string     `json:"pingId" db:"ping_id"`
}

type CreateGeoPing struct {
	Creator   *UserBasic `db:"creator"`
	SentMess  string     `json:"sentMessage" db:"sent_message"`
	Location  *Location  `json:"location" db:"location"`
	IsPrivate bool       `json:"isPrivate" db:"is_private"`
	TimeLimit int64      `json:"timeLimit" db:"time_limit"`
}
