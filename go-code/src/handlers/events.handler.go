package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func DeleteEvent(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Event ID",
			"data":  nil,
		})
		return
	}

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
			"MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events {event_id: $event_id}) DETACH DELETE event",
			gin.H{
				"uid":      uid,
				"event_id": c.Param("id"),
			},
		)
		if err != nil {
			return nil, err
		}
		return nil, nil
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
			"data":  "Event has been deleted",
		})
		return
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
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Event ID",
			"data":  nil,
		})
		return
	}

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

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (creator:User)-[:CREATED]->(event:Events{event_id: $event_id}) WHERE event.isPrivate=false OR (event)-[:INVITED]->(:User {user_id: $uid}) "+
				"RETURN event.name, event.rating, event.startTime, event.endTime, event.type, "+
				"event.position, event.description, event.isPrivate, creator.user_id, creator.name",
			gin.H{
				"uid":      uid,
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

func GetUserCreatedEvents(c *gin.Context) {
	if c.Query("offset") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Offset",
			"data":  nil,
		})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	if c.Query("limit") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Limit",
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	if c.Query("userCreated") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing UID of Creator",
			"data":  nil,
		})
		return
	}

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

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		query := ""
		if c.Query("userCreated") == uid {
			query = "MATCH (userA:User {user_id: $user_id})-[:CREATED]->(event:Events)" +
				"RETURN event.event_id, event.name, event.type, event.isPrivate ORDER BY event.startTime " +
				"DESC SKIP $offset LIMIT $limit;"
		} else {
			query = "MATCH (userA:User {user_id: $user_id})-[:CREATED]->(event:Events)" +
				"WHERE event.isPrivate=false OR (event)-[:INVITED]->(:User {user_id: $uid})" +
				"RETURN event.event_id, event.name, event.type, event.isPrivate ORDER BY event.startTime " +
				"DESC SKIP $offset LIMIT $limit;"
		}

		data, err := transaction.Run(
			query, gin.H{
				"user_id": c.Query("userCreated"),
				"my_id":   uid,
				"offset":  offset,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]interface{}, 0)
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

func UpdateEvent(c *gin.Context) {
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

	if jsonData["id"] = c.Param("id"); jsonData["id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Event ID",
			"data":  nil,
		})
		return
	}

	jsonData["uid"] = uid

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events {event_id: $id}) SET event.name=$name, event.startTime=datetime($startTime), "+
				"event.endTime=datetime($endTime), event.type=$type, event.position=point({latitude: $location.latitude, longitude: $location.longitude}), "+
				"event.description=$description, event.isPrivate=$isPrivate; MATCH (event:Events {event_id: $id})-[i:INVITED]->(:Users) DELETE i",
			jsonData,
		)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Interval Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Event updated",
	})
	return
}

func CreateEvent(c *gin.Context) {
	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	jsonData["creator"] = uid

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	d, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User{user_id: $creator}) MERGE (userA)-[:CREATED]->(event:Events "+
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
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  d,
	})
	return
}

func GetAttendees(c *gin.Context) {
	if c.Query("offset") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Offset",
			"data":  nil,
		})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	if c.Query("limit") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Limit",
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

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

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userA:User)-[a:ATTENDED]->(event:Events {event_id: $event_id}) WHERE any(uid IN userA.user_id WHERE uid = $uid)"+
				"RETURN userA.user_id AS id, userA.name AS name, userA.bio AS bio, userA.profilepic AS profilepic ORDER BY a.timeAttended "+
				"SKIP $offset LIMIT $limit;",
			gin.H{
				"uid":      uid,
				"event_id": c.Param("id"),
				"offset":   offset,
				"limit":    limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]interface{}, 0)
		for data.Next() {
			records = append(records, gin.H{
				"id":         ValueExtractor(data.Record().Get("id")).(string),
				"name":       ValueExtractor(data.Record().Get("name")).(string),
				"profilepic": ValueExtractor(data.Record().Get("profilepic")).(bool),
				"bio":        ValueExtractor(data.Record().Get("bio")).(string),
			})
		}
		return records, data.Err()
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

func checkOut(context *gin.Context) {
	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(context.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	jsonData["event_id"] = context.Param("id")

	uid, exists := context.Get("uid")
	if !exists {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}
	jsonData["user_id"] = uid

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

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
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Checked out Successfully",
	})
}

func checkIn(context *gin.Context) {
	uid, exists := context.Get("uid")
	if !exists {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id}) MATCH (event:Events {event_id: $event_id}) WHERE event.isPrivate=false OR (event)-[:INVITED]->(user)"+
				"MERGE (userA)-[:ATTENDED {timeAttended: datetime(), rating: 3, review: ''}]->(event) SET userA.isCheckedIn=true",
			gin.H{
				"user_id":  uid,
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
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Checked in Successfully",
	})
}

func ShareEvent(c *gin.Context) {
	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}
	if len(jsonData["ids"].([]string)) > 30 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Too Many Members",
			"data":  nil,
		})
		return
	}
	jsonData["event_id"] = c.Param("id")

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	jsonData["uid"] = uid

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events {event_id: $event_id})-[i:INVITED]->(u:User) DELETE i;",
			jsonData,
		)
		if err != nil {
			return false, err
		}
		_, err = transaction.Run(
			"UNWIND $ids AS invitee MATCH (user:User {user_id: invitee}) MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events {event_id: $event_id})"+
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

func GetPrivateEventShares(c *gin.Context) {
	if c.Query("offset") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Offset",
			"data":  nil,
		})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	if c.Query("limit") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Limit",
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}
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

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(event:Events{event_id: $event_id})-[:INVITED]->(users:User) "+
				"RETURN users.user_id, users.name, users.profilepic, users.bio "+
				"DESC SKIP $offset LIMIT $limit;",
			gin.H{
				"uid":      uid,
				"event_id": c.Query("userCreated"),
				"offset":   offset,
				"limit":    limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]interface{}, 0)
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

func EndEvent(c *gin.Context) {
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

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (:User {user_id: $uid})-[:CREATED]->(e:Events {event_id: $id}) SET e.isEnded=true MATCH (u:User)-[a:ATTENDED]->(e)"+
				"SET a.timeExited=timestamp(), u.isCheckedIn=false RETURN u.user_id AS uid",
			gin.H{
				"id":  c.Param("id"),
				"uid": uid,
			},
		)

		if err != nil {
			return nil, err
		}

		records := make([]string, 0)
		recordData := record.Record()
		for record.Next() {
			records = append(records, ValueExtractor(recordData.Get("uid")).(string))
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
	//Send ping at event end
	_ = data

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  nil,
	})
	return
}

func EventCleaner() {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"CALL apoc.period.schedule('event-cleaner', 'MATCH (e:Events)"+
				"WHERE e.endTime <= timestamp()"+
				"SET e.isEnded=true MATCH (u:User)-[a:ATTENDED]->(e) SET a.timeExited=timestamp(), u.isCheckedIn=false "+
				"RETURN u.user_id AS uid, e.name AS eventName', 60)\nYIELD uid, eventName\nRETURN uid, eventName",
			gin.H{},
		)
		if err != nil {
			return nil, err
		}
		if record.Err() != nil {
			return nil, record.Err()
		}
		return record.Record(), nil
	})

	_ = data
	if err != nil {
		panic(err.Error())
	}
}
