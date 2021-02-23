package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"io"
	"net/http"
	"pingserver/src/broker"
	dbclient "pingserver/src/db_client"
)

var markersManager = broker.NewRoomManager()
var eventManager = broker.NewRoomManager()

func GetUserBasic(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	//TODO add rest of query
	_, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) RETURN userA.name",
			map[string]interface{}{
				"uid": c.Param("uid"),
			})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values(), nil
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"isEmpty": true,
			"data": nil,
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"isEmpty": false,
			"data": map[string]interface{}{
				//"name": ,
				//"bio": ,
				//"profilepic": ,
			},
		})
	}
}

func GetUserSocials(c *gin.Context) {

}

func CheckinUser(c *gin.Context){
	eventManager.Submit(c.PostForm("eventId"), "attendee-added")
}

func CheckoutUser(c *gin.Context) {
	eventManager.Submit(c.PostForm("eventId"), "attendee-removed")
}

func CreateNewUser(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	input := map[string]interface{}{
		"uid": c.Param("uid"),
		"name": c.PostForm("name"),
		"number": c.PostForm("number"),
	}

	//TODO add rest of query
	_, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MERGE (userA:User {user_id:$uid, name:$name})-[:DIGITAL_PROFILE]->" +
				"(social:Social{number:$number})",
			input)
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record(), nil
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"message": c.PostForm("name") + " has been created",
		})
	}
}

func UpdateUserInfo(c *gin.Context) {

}

func GetRelevantMarkers(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	roomid := c.Param("uid")
	listener := markersManager.OpenListener(roomid)
	defer markersManager.CloseListener(roomid, listener)
	defer markersManager.DeleteBroadcast(roomid)

	clientGone := c.Writer.CloseNotify()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case message := <-listener:
			c.SSEvent("message", message)
			return true
		}
	})

}
