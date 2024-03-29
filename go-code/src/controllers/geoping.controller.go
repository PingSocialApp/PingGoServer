package controllers

import (
	"log"
	"net/http"
	dbclient "pingserver/db_client"
	firebase "pingserver/firebase_client"
	"pingserver/models"
	"pingserver/queue"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func ShareGeoPing(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData models.ShareGeoPing
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	jsonData.PingID = c.Param("id")

	jsonData.Creator = &models.UserBasic{
		UID: uid.(string),
	}

	output, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		records, err := transaction.Run(
			"UNWIND $ids AS invitee MATCH (user:User {user_id: invitee}) MATCH (creator:User {user_id: $creator.uid})-[:CREATED]->(ping:GeoPing {ping_id: $ping_id})"+
				"MERGE (user)-[:VIEWER]->(ping) RETURN user.notifToken AS notifToken,creator.name AS creatorName,ping.sentMessage AS content",
			structToDbMap(jsonData),
		)
		if err != nil {
			return nil, err
		}

		data := gin.H{
			"notifToken":  make([]string, 0),
			"creatorName": "",
			"content":     "",
		}

		for records.Next() {
			record := records.Record()

			data["creatorName"] = ValueExtractor(record.Get("creatorName")).(string)
			data["content"] = ValueExtractor(record.Get("content")).(string)
			if ValueExtractor(record.Get("notifToken")).(string) != "" {
				data["notifToken"] = append(data["notifToken"].([]string), ValueExtractor(record.Get("notifToken")).(string))
			}
		}

		return data, records.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		log.Println(err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  "GeoPing successfully Shared",
		})

		queue.Dispatcher.Dispatch(func() {
			dataPackage := output.(gin.H)

			firebase.SendMultiNotif(dataPackage["notifToken"].([]string), &firebase.Message{
				Title: "New GeoPing! 🌎",
				Data: map[string]string{
					"entity": "geoping",
					"id":     jsonData.PingID,
				},
				Body: dataPackage["creatorName"].(string) + " : " + dataPackage["content"].(string),
			})
		})
		return
	}
}

func CreateGeoPing(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	var jsonData models.CreateGeoPing
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	jsonData.Creator = &models.UserBasic{
		UID: uid.(string),
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	ret, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User {user_id: $creator.uid})"+
				"CREATE (userA)-[:CREATED]->(geoPing:GeoPing {ping_id: apoc.create.uuid(), sentMessage:$sent_message, "+
				"timeCreate: datetime(), timeExpire: datetime()+duration({minutes: $time_limit}),position: point({latitude: $location.latitude, longitude: $location.longitude}), "+
				"isPrivate:$is_private}) RETURN geoPing.ping_id",
			structToDbMap(jsonData),
		)
		if err != nil {
			return false, err
		}
		if record.Next() {
			return gin.H{
				"id": ValueExtractor(record.Record().Get("geoPing.ping_id")),
			}, nil
		}
		return nil, record.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  ret,
	})
}

func DeleteGeoPing(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(geoPing:GeoPing {ping_id: $ping_id}) DETACH DELETE geoPing",
			gin.H{
				"uid":     uid,
				"ping_id": c.Param("id"),
			},
		)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		log.Println(err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  "GeoPing successfully Shared",
		})
		return
	}
}
