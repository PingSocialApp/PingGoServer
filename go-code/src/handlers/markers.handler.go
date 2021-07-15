package handlers

import (
	"fmt"
	"net/http"
	dbclient "pingserver/db_client"
	"pingserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
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

	latitude, err := strconv.ParseFloat(c.Query("latitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Latitude Parameter",
			"data":  nil,
		})
		return
	}
	longitude, err := strconv.ParseFloat(c.Query("longitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Longitude Parameter",
			"data":  nil,
		})
		return
	}
	radius, err := strconv.ParseFloat(c.Query("radius"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Radius Parameter",
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User)-[:CREATED]->(geoPing:GeoPing) WHERE ((userA.user_id = $user_id) OR (geoPing.isPrivate = false)) "+
				"AND (distance(geoPing.position, point({latitude:$latitude, longitude: $longitude})) <= $radius) "+
				"RETURN DISTINCT geoPing.sentMessage, geoPing.isPrivate, geoPing.position AS position, geoPing.timeCreate, geoPing.ping_id, userA.name, userA.profilepic"+
				" ORDER BY position "+
				"UNION "+
				"MATCH (user:User {user_id: $user_id})-[:VIEWER]->(geoPing:GeoPing)<-[:CREATED]-(userA:User)"+
				" WHERE (distance(geoPing.position, point({latitude:$latitude, longitude: $longitude})) <= $radius) "+
				"RETURN DISTINCT geoPing.sentMessage, geoPing.isPrivate, geoPing.position AS position, geoPing.timeCreate, geoPing.ping_id, userA.name, userA.profilepic"+
				" ORDER BY position",
			gin.H{
				"uid":       uid,
				"latitude":  latitude,
				"longitude": longitude,
				"radius":    radius,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]*models.GeoJson, 0)
		for record.Next() {
			recordRaw := record.Record()
			point := ValueExtractor(recordRaw.Get("position")).(neo4j.Point2D)
			records = append(records, &models.GeoJson{
				Properties: &models.GeoPingProp{
					ID:          ValueExtractor(recordRaw.Get("gepPing.ping_id")).(string),
					SentMessage: ValueExtractor(recordRaw.Get("geoPing.sentMessage")).(string),
					IsPrivate:   ValueExtractor(recordRaw.Get("geoPing.isPrivate")).(bool),
					TimeCreate:  ValueExtractor(recordRaw.Get("geoPing.timeCreate")).(time.Time).UTC(),
					Creator: &models.User{
						Name:       ValueExtractor(recordRaw.Get("userA.name")).(string),
						ProfilePic: ValueExtractor(recordRaw.Get("userA.profilePic")).(string),
					},
				},
				Geometry: models.GetNewGeometry(point.X, point.Y),
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
		"data": gin.H{
			"type":     "FeatureCollection",
			"features": structToJsonMap(data),
		},
	})
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

	latitude, err := strconv.ParseFloat(c.Query("latitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Latitude Parameter",
			"data":  nil,
		})
		return
	}
	longitude, err := strconv.ParseFloat(c.Query("longitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Longitude Parameter",
			"data":  nil,
		})
		return
	}
	radius, err := strconv.ParseFloat(c.Query("radius"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Radius Parameter",
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (host:User)-[:CREATED]->(event:Events)"+
				" WHERE ((host.user_id = $user_id) OR (event.isPrivate = false)) AND (distance(event.position, point({latitude: $latitude, longitude: $longitude})) <= $maxDistance) AND (event.isEnded = false) AND event.startTime <= (datetime() + duration('P1D'))"+
				" RETURN DISTINCT event.event_id, event.type, event.name, event.isPrivate, event.rating, event.startTime, event.endTime, event.position AS position, host.name, host.profilepic, host.user_id"+
				" ORDER BY position"+
				" UNION "+
				"MATCH (host:User)-[:CREATED]->(event:Events)-[:INVITED]->(invitee:User {user_id: $user_id})"+
				" WHERE (distance(event.position, point({latitude: $latitude, longitude: $longitude})) <= $maxDistance) AND (event.isEnded = false) AND event.startTime <= (datetime() + duration('P1D'))"+
				" RETURN DISTINCT event.event_id, event.type, event.name, event.isPrivate, event.rating, event.startTime, event.endTime, event.position AS position, host.name, host.profilepic, host.user_id"+
				" ORDER BY position;",
			gin.H{
				"user_id":     uid,
				"latitude":    latitude,
				"longitude":   longitude,
				"maxDistance": radius,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]*models.GeoJson, 0)
		for record.Next() {
			recordRaw := record.Record()
			point := ValueExtractor(recordRaw.Get("position")).(neo4j.Point2D)
			records = append(records, &models.GeoJson{
				Properties: &models.EventProp{
					ID: ValueExtractor(recordRaw.Get("event.event_id")).(string),
					Creator: &models.User{
						Name:       ValueExtractor(recordRaw.Get("host.name")).(string),
						ProfilePic: ValueExtractor(recordRaw.Get("host.profilePic")).(string),
						ID:         ValueExtractor(recordRaw.Get("host.user_id")).(string),
					},
					Name:      ValueExtractor(recordRaw.Get("event.name")).(string),
					Type:      ValueExtractor(recordRaw.Get("event.type")).(string),
					IsPrivate: ValueExtractor(recordRaw.Get("event.isPrivate")).(bool),
					Rating:    ValueExtractor(recordRaw.Get("event.rating")).(int64),
					StartTime: ValueExtractor(recordRaw.Get("event.startTime")).(time.Time).UTC(),
					EndTime:   ValueExtractor(recordRaw.Get("event.endTime")).(time.Time).UTC(),
				},
				Geometry: models.GetNewGeometry(point.X, point.Y),
			})
		}
		return records, record.Err()
	})

	if err != nil {
		fmt.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		return
	}

	outputData := data.([]*models.GeoJson)
	outputArr := make(gin.H, len(outputData))

	for index, element := range outputArr {
		outputArr[index] = structToJsonMap(element)
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data": gin.H{
			"type":     "FeatureCollection",
			"features": outputArr,
		},
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
				" WHERE link.permissions >= 2048 AND userA.isCheckedIn=false AND distance(userA.location,point({latitude: $position.latitude, longitude: $position.longitude})) <= $radius"+
				" RETURN userA.name AS name, userA.user_id AS id, userA.profilepic AS profilepic, userA.bio AS bio, userA.location AS location",
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
		records := make([]*models.GeoJson, 0)
		for record.Next() {
			recordRaw := record.Record()
			point := ValueExtractor(recordRaw.Get("location")).(neo4j.Point2D)
			records = append(records, &models.GeoJson{
				Properties: &models.User{
					ID:         ValueExtractor(recordRaw.Get("id")).(string),
					Name:       ValueExtractor(recordRaw.Get("name")).(string),
					Bio:        ValueExtractor(recordRaw.Get("bio")).(string),
					ProfilePic: ValueExtractor(recordRaw.Get("profilePic")).(string),
				},
				Geometry: models.GetNewGeometry(point.X, point.Y),
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
		"data": gin.H{
			"type":     "FeatureCollection",
			"features": data,
		},
	})
	return
}
