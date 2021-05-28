package dbclient

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

var DB neo4j.Driver

func CreateDriver(uri, username, password string) {
	db, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		panic(err.Error())
	}else{
		DB = db
	}

	err = db.VerifyConnectivity()
	if err != nil {
		panic(err.Error())
	}
}

func CloseDriver() {
	err := DB.Close()
	if err != nil {
		panic(err.Error())
	}
}

func CreateSession() (neo4j.Session){	
	session, err := DB.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	if err != nil {
		panic(err.Error())
	} else {
		return session
	}
}

func KillSession(session neo4j.Session) {
	err := session.Close()
	if err != nil {
		panic(err.Error())
	}
}
