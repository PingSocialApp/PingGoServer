package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"
)

func ShareGeoPing(c *gin.Context){
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H  // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   e.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	jsonData["ping_id"] = c.Param("id")

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
				"UNWIND $ids AS invitee MATCH (user:User {user_id: invitee}) MATCH (ping:GeoPing {ping_id: $ping_id})" +
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
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error":   nil,
			"isEmpty": false,
			"data":    "Event successfully Shared",
		})
	}
}

func CreateGeoPing(c* gin.Context){
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H  // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   e.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}

	ret, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id})" +
				"CREATE (userA)-[:CREATED]->(geoPing:GeoPing {ping_id: apoc.create.uuid(), sentMessage:$sentMessage, " +
				"timeCreate: datetime(), position: point({latitude: $position.latitude, longitude: $position.longitude}), " +
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
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"isEmpty": false,
		"data":  ret,
	})
}

func DeleteGeoPing(c *gin.Context){
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (geoPing:GeoPing {ping_id: $ping_id}) DETACH DELETE geoPing;",
			gin.H{
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
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error":   nil,
			"isEmpty": false,
			"data":    "GeoPing successfully Shared",
		})
	}
}