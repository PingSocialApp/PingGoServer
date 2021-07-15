package models

type ShareGeoPing struct {
	ID     []string `json:"ids" db:"ids"`
	UID    string   `db:"uid"`
	PingID string   `json:"pingId" db:"ping_id"`
}

type CreateGeoPing struct {
	UID       string    `db:"user_id"`
	SentMess  string    `json:"sentMessage" db:"sent_message"`
	Location  *Location `json:"location" db:"position"`
	IsPrivate bool      `json:"isPrivate" db:"is_private"`
	TimeLimit int64     `json:"timeLimit" db:"time_limit"`
}
