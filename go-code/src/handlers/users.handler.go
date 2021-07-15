package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"
	"pingserver/models"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"github.com/gin-gonic/gin"
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
			"MATCH (userA:User {user_id: $uid}) RETURN userA.name, userA.bio, userA.profilepic, userA.checkedIn",
			gin.H{
				"uid": c.Param("uid"),
			},
		)
		if err != nil {
			return nil, err
		}
		if result.Next() {
			data := result.Record()
			user := &models.UserBasic{
				UID:        c.Param("uid"),
				Bio:        ValueExtractor(data.Get("userA.bio")).(string),
				ProfilePic: ValueExtractor(data.Get("userA.profilepic")).(string),
				Name:       ValueExtractor(data.Get("userA.name")).(string),
				CheckedIn:  ValueExtractor(data.Get("userA.checkedIn")).(string),
			}

			return user, nil
		}

		return nil, result.Err()
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
			"data":  transaction,
		})
		return
	}
}

func CreateNewUser(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData models.UserBasic // map[string]interface{}
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	uid, exists := c.Get("uid")
	if exists {
		jsonData.UID = uid.(string)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"CREATE (userA:User {user_id:$uid, name:$name, bio:$bio, profilepic:$profile_pic, checkedIn:''})",
			structToDbMap(jsonData))
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
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  jsonData.Name + " has been created",
	})
	return
}

func UpdateUserInfo(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData models.UserBasic // map[string]interface{}
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	uid, exists := c.Get("uid")
	if exists {
		jsonData.UID = uid.(string)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) SET userA.name=$name, userA.bio=$bio, userA.profilepic=$profile_pic",
			structToDbMap(jsonData))
		if err != nil {
			return nil, err
		}

		return nil, result.Err()
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
		"data":  jsonData.Name + " has been updated",
	})
	return
}

func SetUserLocation(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData models.UserBasic // map[string]interface{}
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshal error
			"data":  nil,
		})
		return
	}
	uid, exists := c.Get("uid")
	if exists {
		jsonData.UID = uid.(string)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}
	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid})-[a:ATTENDED]->(:Event) WHERE userA.checkedIn='' "+
				"SET userA.location = point({latitude: $location.latitude, longitude: $location.longitude})",
			structToDbMap(jsonData))
		if err != nil {
			return nil, err
		}

		return nil, result.Err()
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
		"data":  "Location has been updated",
	})
	return
}

func GetUserLocation(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	locationData, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		isCheckedInRecord, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id) RETURN userA.checkedIn='' AS isCheckedIn", gin.H{
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
		point := ValueExtractor(record.Get("location")).(neo4j.Point2D)
		if locationRecord.Next() {
			return &models.Location{
				Latitude:  point.X,
				Longitude: point.Y,
			}, nil
		}

		return nil, locationRecord.Err()
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
			"data":  locationData,
		})
		return
	}
}

func SetNotifToken(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	var jsonData gin.H // map[string]interface{}
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshal error
			"data":  nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
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
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  "Token successfully updated",
		})
		return
	}

}
