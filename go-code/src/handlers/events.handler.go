package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func DeleteEvent(c *gin.Context) {
	if c.Param("id") == ""{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Event ID",
			"data": nil,
		})
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid User",
			"data": nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	successful, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		//TODO Return t/f based on if successful delete
		successful, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events {event_id: $event_id}) DETACH DELETE event",
			gin.H{
				"uid": uid,
				"event_id": c.Param("id"),
			},
		)
		if err != nil {
			return nil, err
		} else if successful.Next(){
			return successful.Record(), err
		}
		return nil, successful.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"data":    nil,
		})
	} else if successful.(bool){
		c.JSON(http.StatusOK, gin.H{
			"error":   nil,
			"data":    "Event has been deleted",
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   nil,
			"data":    "Event has been deleted",
		})
	}

}

func HandleAttendance(c *gin.Context) {
	if c.Query("action") == "checkout" {
		checkIn(c.Copy())
	} else {
		checkOut(c.Copy())
	}
}

func GetEventDetails(c *gin.Context) {
	if c.Param("id") == ""{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Event ID",
			"data": nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (creator:User)-[:CREATED]->(event:Events{event_id: $event_id})"+
				"RETURN event.name, event.rating, event.startTime, event.endTime, event.type, "+
				"event.position, event.description, event.isPrivate, creator.user_id, creator.name",
			gin.H{
				"event_id": c.Param("id"),
			},
		)
		if err != nil {
			return nil, err
		}

		if result.Next() {
			data := result.Record()
			point := ValueExtractor(data.Get("event.position")).(*neo4j.Point)
			return gin.H{
				"eventName":   ValueExtractor(data.Get("event.name")).(string),
				"description": ValueExtractor(data.Get("event.description")).(string),
				"type":        ValueExtractor(data.Get("event.type")).(string),
				"position": gin.H{
					"latitude":  point.X(),
					"longitude": point.Y(),
				},
				"rating":      ValueExtractor(data.Get("event.rating")).(float64),
				"isPrivate":   ValueExtractor(data.Get("event.isPrivate")).(bool),
				"startTime":   ValueExtractor(data.Get("event.startTime")).(time.Time).Format(time.RFC3339),
				"endTime":     ValueExtractor(data.Get("event.endTime")).(time.Time).Format(time.RFC3339),
				"creatorId":   ValueExtractor(data.Get("creator.user_id")).(string),
				"creatorName": ValueExtractor(data.Get("creator.name")).(string),
			}, nil
		}

		return nil, result.Err()
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
		"data":    data,
	})
}

func GetUserCreatedEvents(c *gin.Context) {
	//TODO Convert myId to session
	offset := 0
	limit := 50

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"data":    nil,
		})
		return
	}
	limit, err = strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"data":    nil,
		})
		return
	}
	if c.Query("userCreated") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing UID of Creator",
			"data":    nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id})-[:CREATED]->(event:Events)"+
				"WHERE event.isPrivate=false "+
				"RETURN event.event_id, event.name, event.type, event.isPrivate ORDER BY event.startTime "+
				"DESC SKIP $offset LIMIT $limit;",
			gin.H{
				"user_id": c.Query("userCreated"),
				"my_id":   c.Query("myId"),
				"offset":  offset,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}
		var records []interface{}
		for data.Next() {
			records = append(records, gin.H{
				"id":        ValueExtractor(data.Record().Get("event.event_id")).(string),
				"name":      ValueExtractor(data.Record().Get("event.name")).(string),
				"isPrivate": ValueExtractor(data.Record().Get("event.isPrivate")).(bool),
				"type":      ValueExtractor(data.Record().Get("event.type")).(string),
			})
		}
		return records, data.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"data":    data.(interface{}),
	})

}

func UpdateEvent(c *gin.Context) {
	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   e.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}

	if jsonData["id"] = c.Param("id"); jsonData["id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing Event ID",
			"data":    nil,
		})
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events {event_id: $id}) SET event.name=$name, event.startTime=datetime($startTime), "+
				"event.endTime=datetime($endTime), event.type=$type, event.position=point({latitude: $location.latitude, longitude: $location.longitude}), "+
				"event.description=$description, event.isPrivate=$isPrivate",
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
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"data": "Event updated",
	})
}

func CreateEvent(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   e.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}

	d, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User{user_id: $Creator}) MERGE (userA)-[:CREATED]->(event:Events "+
				"{event_id: apoc.create.uuid(), name: $name, rating: 3.0, startTime: datetime($startTime), "+
				"endTime: datetime($endTime), isEnded:false, type: $type, position: point({latitude: $location.latitude, longitude: $location.longitude}), "+
				"description: $description, isPrivate: $isPrivate }) RETURN event.event_id AS eventId",
			jsonData,
		)
		if err != nil {
			return false, err
		}
		if record.Next() {
			return gin.H{
				"id": ValueExtractor(record.Record().Get("eventId")),
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
		"data":    d,
	})
}

func GetAttendees(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userA:User)-[a:ATTENDED]->(event:Events {event_id: $event_id}) "+
				"RETURN userA.user_id AS id, userA.name AS name, userA.bio AS bio, userA.profilepic AS profilepic ORDER BY a.timeAttended "+
				"SKIP $offset LIMIT $limit;",
			gin.H{
				"event_id": c.Param("id"),
				"offset":  offset,
				"limit":    limit,
			},
		)
		if err != nil {
			return nil, err
		}
		var records []interface{}
		for data.Next() {
			records = append(records, gin.H{
				"id":        ValueExtractor(data.Record().Get("id")).(string),
				"name":      ValueExtractor(data.Record().Get("name")).(string),
				"profilepic": ValueExtractor(data.Record().Get("profilepic")).(bool),
				"bio":      ValueExtractor(data.Record().Get("bio")).(string),
			})
		}
		return records, data.Err()
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
		"isEmpty": len(data.([]interface{})) == 0,
		"data":    data,
	})
}

func checkOut(context *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(context.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error":   e.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	jsonData["event_id"] = context.Param("id")

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id})-[a:ATTENDED]->(event:Events {event_id: $event_id})"+
				"SET a.timeExited = datetime(), a.rating = $rating, a.review = $review, userA.isCheckedIn=false;"+
				"MATCH (:User)-[a:ATTENDED]->(event:Events {event_id: $event_id}) SET event.rating = avg(a.rating)",
			jsonData,
		)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"isEmpty": false,
		"data":    "Checked out Successfully",
	})
}

func checkIn(context *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id}) MATCH (event:Events {event_id: $event_id}) "+
				"MERGE (userA)-[:ATTENDED {timeAttended: datetime(), rating: 3, review: ''}]->(event) SET userA.isCheckedIn=true",
			gin.H{
				"user_id":  context.Query("uid"),
				"event_id": context.Param("id"),
			},
		)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"isEmpty": false,
		"data":    "Checked in Successfully",
	})
}

func ShareEvent(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   e.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	if len(jsonData["ids"].([]interface{})) > 30 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Too Many Memebers",
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	jsonData["event_id"] = c.Param("id")

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (event:Events {event_id: $event_id})-[i:INVITED]->(u:User) DELETE i;",
			jsonData,
		)
		if err != nil {
			return false, err
		}
		_, err = transaction.Run(
			"UNWIND $ids AS invitee MATCH (user:User {user_id: invitee}) MATCH (event:Events {event_id: $event_id})"+
				"MERGE (event)-[:INVITED]->(user);",
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

func GetPrivateEventShares(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (event:Events{event_id: $event_id})-[:INVITED]->(users:User) "+
				"RETURN users.user_id, users.name, users.profilepic, users.bio "+
				"DESC SKIP $offset LIMIT $limit;",
			gin.H{
				"event_id": c.Query("userCreated"),
				"offset":   offset,
				"limit":    limit,
			},
		)
		if err != nil {
			return nil, err
		}
		var records []interface{}
		for data.Next() {
			records = append(records, gin.H{
				"id":         ValueExtractor(data.Record().Get("users.user_id")).(string),
				"name":       ValueExtractor(data.Record().Get("users.name")).(string),
				"bio":        ValueExtractor(data.Record().Get("users.bio")).(string),
				"profilepic": ValueExtractor(data.Record().Get("users.profilepic")).(string),
			})
		}
		return records, data.Err()
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
		"isEmpty": len(data.([]interface{})) == 0,
		"data":    data.([]interface{}),
	})
}

func EndEvent(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (e:Events {event_id: $id}) SET e.isEnded=true MATCH (u:User)-[a:ATTENDED]->(e)"+
				"SET a.timeExited=timestamp(), u.isCheckedIn=false RETURN u.user_id AS uid",
			gin.H{
				"id": c.Param("id"),
			},
		)

		if err != nil {
			return nil, err
		}

		var records []string
		recordData := record.Record()
		for record.Next() {
			records = append(records, ValueExtractor(recordData.Get("uid")).(string))
		}
		return records, record.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"isEmpty": true,
			"data":    nil,
		})
		return
	}
	//Send ping at event end
	_ = data

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"isEmpty": false,
		"data":    nil,
	})
}

func EventCleaner(){
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"CALL apoc.period.schedule('event-cleaner', 'MATCH (e:Events)" +
				"WHERE e.endTime <= timestamp()" +
				"SET e.isEnded=true MATCH (u:User)-[a:ATTENDED]->(e) SET a.timeExited=timestamp(), u.isCheckedIn=false " +
				"RETURN u.user_id AS uid, e.name AS eventName', 60)\nYIELD uid, eventName\nRETURN uid, eventName",
			gin.H{
			},
		)
		if err != nil {
			return nil, err
		}
		if record.Err() != nil{
			return nil, record.Err()
		}
		return record.Record(), nil
	})

	_=data
	if err != nil {
		panic(err.Error())
	}
}
