package models

type Response struct {
	Error error       `json:"error"`
	Data  interface{} `json:"data"`
}
