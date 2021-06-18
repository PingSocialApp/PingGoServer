package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func ValueExtractor(data interface{}, exists bool) (ret interface{}) {
	if exists {
		return data
	} else {
		return nil
	}
}

func GetUserBasic(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	transaction, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $UID}) RETURN userA.name, userA.bio, userA.profilepic, userA.isCheckedIn",
			gin.H{
				"UID": c.Param("uid"),
			},
		)
		if err != nil {
			return nil, err
		}
		if result.Next() {
			data := result.Record()
			user := gin.H{
				"bio":         ValueExtractor(data.Get("userA.bio")).(string),
				"profilepic":  ValueExtractor(data.Get("userA.profilepic")).(string),
				"name":        ValueExtractor(data.Get("userA.name")).(string),
				"isCheckedIn": ValueExtractor(data.Get("userA.isCheckedIn")).(bool),
			}

			return user, nil
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  transaction,
		})
	}
}

func CreateNewUser(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
			"data":  nil,
		})
		return
	}

	uid, exists := c.Get("uid")
	if exists {
		jsonData["uid"] = uid
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MERGE (userA:User {user_id:$uid, name:$name, bio:$bio, profilepic:$profilepic, isCheckedIn:false, "+
				"userType:$userType})",
			jsonData)
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
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  jsonData["name"].(string) + " has been created",
	})
}

func UpdateUserInfo(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
			"data":  nil,
		})
		return
	}

	uid, exists := c.Get("uid")
	if exists {
		jsonData["uid"] = uid
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) SET userA.name=$name, userA.bio=$bio, userA.profilepic=$profilepic",
			jsonData)
		if err != nil {
			return nil, err
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  jsonData["name"].(string) + " has been update",
	})
}

func SetUserLocation(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	uid, exists := c.Get("uid")
	if exists {
		jsonData["uid"] = uid
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}
	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid})-[a:ATTENDED]->(:Event) WHERE userA.isCheckedIn=false SET userA.location = point({latitude: $latitude, longitude: $longitude})",
			jsonData)
		if err != nil {
			return nil, err
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Location has been updated",
	})
}

func GetUserLocation(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	locationData, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		isCheckedInRecord, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id) RETURN userA.isCheckedIn AS isCheckedIn", gin.H{
				"user_id": c.Param("uid"),
			})
		if err != nil {
			return nil, err
		}

		locationRecord, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id,}) CALL apoc.do.when($isCheckedIn, "+
				"'MATCH (userA:User {user_id: $user_id})-[a:ATTENDED]->(e:Event) WHERE a.timeExited IS NULL RETURN e.location AS location', "+
				"'MATCH (userA:User {user_id: $user_id}) RETURN userA.location AS location') YIELD location AS location",
			gin.H{
				"user_id":     c.Param("uid"),
				"isCheckedIn": ValueExtractor(isCheckedInRecord.Record().Get("isCheckedIn")),
			})
		if err != nil {
			return nil, err
		}

		record := locationRecord.Record()
		point := ValueExtractor(record.Get("location")).(*neo4j.Point)
		if locationRecord.Next() {
			return gin.H{
				"latitude":  point.X(),
				"longitude": point.Y(),
			}, nil
		}

		return nil, locationRecord.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  locationData,
		})
	}
}

func SetNotifToken(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) \n userA.notifToken=$token",
			gin.H{
				"uid":   uid,
				"token": c.PostForm("notifToken"),
			})
		if err != nil {
			return nil, err
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  "Token successfully updated",
		})
	}

}
