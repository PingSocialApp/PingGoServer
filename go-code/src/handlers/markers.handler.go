package handlers

import (
	"fmt"
	"net/http"
	dbclient "pingserver/db_client"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func GetGeoPings(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	if c.Query("latitude") == "" || c.Query("longitude") == "" || c.Query("radius") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Parameters",
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User)-[:CREATED]->(geoPing:GeoPing)"+
				"WHERE ((userA.user_id = $user_id) OR (geoPing.isPrivate = false))"+
				"AND (distance(geoPing.position, point({latitude: $latitude, longitude: $longitude})) <= $maxDistance)"+
				"RETURN DISTINCT geoPing.sentMessage, geoPing.isPrivate, geoPing.position AS position, geoPing.timeCreate, geoPing.ping_id, userA.name, userA.profilepic"+
				"ORDER BY position"+
				"UNION"+
				"MATCH (user:User {user_id: $user_id})-[:VIEWER]->(geoPing:GeoPing)<-[:CREATED]-(userA:User)"+
				"WHERE (distance(geoPing.position, point({latitude: $latitude, longitude: $longitude})) <= $maxDistance)"+
				"RETURN DISTINCT geoPing.sentMessage, geoPing.isPrivate, geoPing.position AS position, geoPing.timeCreate, geoPing.ping_id, userA.name, userA.profilepic"+
				"ORDER BY position;",
			gin.H{
				"user_id": uid,
				"position": gin.H{
					"latitude":  c.Query("latitude"),
					"longitude": c.Query("longitude"),
				},
				"radius": c.Query("radius"),
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]interface{}, 0)
		recordRaw := record.Record()
		for record.Next() {
			point := ValueExtractor(recordRaw.Get("position")).(*neo4j.Point)
			records = append(records, gin.H{
				"id":         ValueExtractor(recordRaw.Get("gepPing.ping_id")).(string),
				"message":    ValueExtractor(recordRaw.Get("geoPing.sentMessage")).(string),
				"isPrivate":  ValueExtractor(recordRaw.Get("geoPing.isPrivate")).(bool),
				"timeCreate": ValueExtractor(recordRaw.Get("geoPing.timeCreate")).(time.Time),
				"creator": gin.H{
					"name":       ValueExtractor(recordRaw.Get("userA.name")).(string),
					"profilepic": ValueExtractor(recordRaw.Get("userA.profilepic")).(string),
				},
				"position": gin.H{
					"latitude":  point.X(),
					"longitude": point.Y(),
				},
			})
		}
		return records, record.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  data,
	})
	return
}

func GetEvents(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	if c.Query("latitude") == "" || c.Query("longitude") == "" || c.Query("radius") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Parameters",
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (host:User)-[:CREATED]->(event:Events)"+
				"WHERE ((host.user_id = $user_id) OR (event.isPrivate = false)) AND (distance(event.position, point({latitude: $latitude, longitude: $longitude})) <= $maxDistance) AND (event.isEnded = false) AND event.startTime <= (datetime() + duration('P1D'))"+
				"RETURN DISTINCT event.event_id, event.name, event.isPrivate, event.rating, event.startTime, event.endTime, event.position AS position, host.name, host.profilepic"+
				"ORDER BY position"+
				"UNION"+
				"MATCH (host:User)-[:CREATED]->(event:Events)-[:INVITED]->(invitee:User  {user_id: $user_id})"+
				"WHERE (distance(event.position, point({latitude: $latitude, longitude: $longitude})) <= $maxDistance) AND (event.isEnded = false) AND event.startTime <= (datetime() + duration('P1D'))"+
				"RETURN DISTINCT event.event_id, event.type, event.name, event.isPrivate, event.rating, event.startTime, event.endTime, event.position AS position, host.name, host.profilepic"+
				"ORDER BY position;",
			gin.H{
				"user_id": uid,
				"position": gin.H{
					"latitude":  c.Query("latitude"),
					"longitude": c.Query("longitude"),
				},
				"radius": c.Query("radius"),
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]interface{}, 0)
		recordRaw := record.Record()
		for record.Next() {
			point := ValueExtractor(recordRaw.Get("position")).(*neo4j.Point)
			records = append(records, gin.H{
				"id": ValueExtractor(recordRaw.Get("event.event_id")).(string),
				"creator": gin.H{
					"name":       ValueExtractor(recordRaw.Get("host.name")).(string),
					"profilepic": ValueExtractor(recordRaw.Get("host.profilepic")).(string),
				},
				"type":      ValueExtractor(recordRaw.Get("event.type")).(string),
				"isPrivate": ValueExtractor(recordRaw.Get("event.isPrivate")).(bool),
				"rating":    ValueExtractor(recordRaw.Get("event.rating")).(float64),
				"startTime": ValueExtractor(recordRaw.Get("event.startTime")).(time.Time),
				"endTime":   ValueExtractor(recordRaw.Get("event.endTime")).(time.Time),
				"position": gin.H{
					"latitude":  point.X(),
					"longitude": point.Y(),
				},
			})
		}
		return records, record.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  data,
	})
	return
}

func GetLinkMarkers(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	if c.Query("latitude") == "" || c.Query("longitude") == "" || c.Query("radius") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Parameters",
			"data":  nil,
		})
		return
	}

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User)-[link:LINKED]->(userB:User {user_id: $user_id})"+
				"WHERE link.permissions >= 2048 AND userA.isCheckedIn=false AND distance(userA.location,point({latitude: $position.latitude, longitude: $position.longitude})) <= $radius"+
				"RETURN userA.name AS name, userA.user_id AS id, userA.profilepic AS profilepic, userA.bio AS bio, userA.location AS location",
			gin.H{
				"user_id": uid,
				"position": gin.H{
					"latitude":  c.Query("latitude"),
					"longitude": c.Query("longitude"),
				},
				"radius": c.Query("radius"),
			},
		)

		if err != nil {
			return nil, err
		}
		records := make([]interface{}, 0)
		recordRaw := record.Record()
		for record.Next() {
			point := ValueExtractor(recordRaw.Get("location")).(*neo4j.Point)
			records = append(records, gin.H{
				"id":         ValueExtractor(recordRaw.Get("id")).(string),
				"name":       ValueExtractor(recordRaw.Get("name")).(string),
				"bio":        ValueExtractor(recordRaw.Get("bio")).(string),
				"profilepic": ValueExtractor(recordRaw.Get("profilepic")).(string),
				"position": gin.H{
					"latitude":  point.X(),
					"longitude": point.Y(),
				},
			})
		}
		return records, record.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  data,
	})
	return
}
