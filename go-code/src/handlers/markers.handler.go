package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"net/http"
	"pingserver/broker"
	dbclient "pingserver/db_client"
)
var markersManager = broker.NewRoomManager()
var eventManager = broker.NewRoomManager()
//func GetRelevantMarkers(c *gin.Context) {
//	c.Writer.Header().Set("Content-Type", "text/event-stream")
//	c.Writer.Header().Set("Cache-Control", "no-cache")
//	c.Writer.Header().Set("Connection", "keep-alive")
//	c.Writer.Header().Set("Transfer-Encoding", "chunked")
//
//	roomid := c.Param("uid")
//	listener := markersManager.OpenListener(roomid)
//	defer markersManager.CloseListener(roomid, listener)
//	defer markersManager.DeleteBroadcast(roomid)
//
//	clientGone := c.Writer.CloseNotify()
//	c.Stream(func(w io.Writer) bool {
//		select {
//		case <-clientGone:
//			return false
//		case message := <-listener:
//			c.SSEvent("message", message)
//			return true
//		}
//	})
//
//}

func GetRelevantMarkers(c *gin.Context) {

}

func GetLinkMarkers(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User)-[link:LINKED]->(userB:User {user_id: $user_id})" +
				"WHERE link.permissions >= 2048 AND userA.isCheckedIn=false AND distance(userA.location,point({latitude: $position.latitude, longitude: $position.longitude})) <= $radius" +
				"RETURN userA.name AS name, userA.user_id AS id, userA.profilepic AS profilepic, userA.bio AS bio, userA.location AS location",
			gin.H{
				"user_id": c.Query("uid"),
				"position": gin.H{
					"latitude": c.Query("latitude"),
					"longitude": c.Query("longitude"),
				},
				"radius": c.Query("radius"),
			},
		)

		if err != nil {
			return nil, err
		}
		var records []interface{}
		recordRaw := record.Record()
		for record.Next() {
			point := ValueExtractor(recordRaw.Get("location")).(*neo4j.Point)
			records = append(records, gin.H{
				"id":         ValueExtractor(recordRaw.Get("id")).(string),
				"name":       ValueExtractor(recordRaw.Get("name")).(string),
				"bio":        ValueExtractor(recordRaw.Get("bio")).(string),
				"profilepic": ValueExtractor(recordRaw.Get("profilepic")).(string),
				"position":gin.H {
					"latitude":  point.X(),
					"longitude": point.Y(),
				},
			})
		}
		return records, record.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"isEmpty": len(data.([]interface{})) == 0,
		"data":    data,
	})
}
