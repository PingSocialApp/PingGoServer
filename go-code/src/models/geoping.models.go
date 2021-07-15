package models

type ShareGeoPing struct {
	ID     []string `db:"ids"`
	UID    string   `db:"uid"`
	PingId string   `db:"ping_id"`
}

type CreateGeoPing struct {
	UID      string `db:"user_id"`
	SentMess string `db:"sent_message"`
	// TODO Double check if works in DB
	Location  *Location `db:"position"`
	isPrivate bool      `db:"is_private"`
	timeLimit int64     `db:"time_limit"`
}