package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"

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

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	jsonData["ping_id"] = c.Param("id")
	jsonData["uid"] = uid

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"UNWIND $ids AS invitee MATCH (user:User {user_id: invitee}) MATCH (:User {user_id: $uid})-[:CREATED]->(ping:GeoPing {ping_id: $ping_id})"+
				"MERGE (event)-[:VIEWER]->(user);",
			jsonData,
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
		fmt.Println(err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  "Event successfully Shared",
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

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	jsonData["user_id"] = uid

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	ret, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id})"+
				"CREATE (userA)-[:CREATED]->(geoPing:GeoPing {ping_id: apoc.create.uuid(), sentMessage:$sentMessage, "+
				"timeCreate: datetime(), position: point({latitude: $position.latitude, longitude: $position.longitude}), "+
				"isPrivate:$isPrivate}) WITH geoPing CALL apoc.ttl.expireIn(geoPing, $timeLimit, 'm') WITH geoPing RETURN geoPing.ping_id",
			jsonData,
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
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  ret,
	})
	return
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
		fmt.Println(err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  "GeoPing successfully Shared",
		})
		return
	}
}
