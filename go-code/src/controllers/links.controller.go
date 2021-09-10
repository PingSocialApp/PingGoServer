package controllers

import (
	"errors"
	"log"
	"net/http"
	dbclient "pingserver/db_client"
	firebase "pingserver/firebase_client"
	"pingserver/models"
	"pingserver/queue"
	"strconv"
	"strings"

	"firebase.google.com/go/db"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func DeleteRequest(c *gin.Context) {
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
			"MATCH (:User {user_id: $uid})-[link:REQUESTED {link_id:$lid}]->(:User) DETACH DELETE link;",
			gin.H{
				"uid": uid,
				"lid": c.Param("rid"),
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
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Request Deleted",
	})
}

func AcceptRequest(c *gin.Context) {
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

	output, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			`MATCH (userA:User)-[request:REQUESTED {link_id: $link_id}]->(userB:User {user_id: $uid})
				 MERGE (userA)-[link:LINKED {link_id: request.link_id, permissions: request.permissions}]->(userB)
				 DETACH DELETE request RETURN userB.name AS name, userA.notifToken AS notifToken`,
			gin.H{
				"uid":     uid,
				"link_id": c.Param("rid"),
			},
		)
		if err != nil {
			return nil, err
		}
		if record.Next() {
			return gin.H{
				"notifToken": ValueExtractor(record.Record().Get("notifToken")).(string),
				"name":       ValueExtractor(record.Record().Get("name")).(string),
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
		"data":  "Request Accepted",
	})

	decrementRequestNum(c, uid.(string))

	err = queue.Dispatcher.Dispatch(func() {
		dataPackage := output.(gin.H)

		firebase.SendSingleNotif(dataPackage["notifToken"].(string), &firebase.Message{
			Title: "Request Accepted!",
			Body:  dataPackage["name"].(string) + " has accepted your link request ✅",
		})
	})
	if err != nil {
		log.Println(err.Error())
	}
}

func DeclineRequest(c *gin.Context) {
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
			"MATCH (:User)-[link:REQUESTED {link_id:$lid}]->(:User {user_id: $uid}) DETACH DELETE link",
			gin.H{
				"uid": uid,
				"lid": c.Param("rid"),
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
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Request Deleted",
	})

	decrementRequestNum(c, uid.(string))
}

func decrementRequestNum(c *gin.Context, uid string) {
	ref := firebase.RTDB.NewRef("userNumerics/numRequests/" + uid)

	fn := func(t db.TransactionNode) (interface{}, error) {
		var currentValue int
		if err := t.Unmarshal(&currentValue); err != nil {
			return currentValue, err
		}
		if currentValue <= 0 {
			return 0, nil
		}
		return currentValue - 1, nil
	}

	if err := ref.Transaction(c, fn); err != nil {
		log.Println("Transaction failed to commit:", err)
	}
}

func SendRequest(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	var jsonData models.Request
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	jsonData.Me = &models.UserBasic{
		UID: uid.(string),
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	output, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		inputs := structToDbMap(jsonData)

		exists, err := transaction.Run(
			"MATCH (userA:User {user_id: $me.uid}) MATCH (userB:User {user_id: $user_rec.uid}) "+
				"RETURN (exists((userA)-[:REQUESTED]->(userB)) OR exists((userA)-[:LINKED]->(userB))) AS linkExists",
			inputs,
		)
		if err != nil {
			return nil, err
		}

		if exists.Next() && ValueExtractor(exists.Record().Get("linkExists")).(bool) {
			return "exists", nil
		} else if exists.Err() != nil {
			return nil, err
		}

		data, err := transaction.Run(
			`MATCH (userA:User {user_id: $me.uid}) MATCH (userB:User {user_id: $user_rec.uid}) 
			CREATE (userA)-[r:REQUESTED {link_id: apoc.create.uuid(), permissions: $permissions}]->(userB) RETURN userA.name AS sendName, userB.notifToken AS notifToken`,
			inputs,
		)

		if err != nil {
			return nil, err
		}

		if data.Next() {
			return gin.H{
				"notifToken": ValueExtractor(data.Record().Get("notifToken")).(string),
				"name":       ValueExtractor(data.Record().Get("sendName")).(string),
			}, nil
		}
		return nil, data.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		log.Println(err.Error())
		return
	}

	switch output {
	case "exists":
		c.JSON(http.StatusConflict, gin.H{
			"error": nil,
			"data":  "Record already exists",
		})
		return
	default:
		ref := firebase.RTDB.NewRef("userNumerics/numRequests/" + jsonData.UserRec.UID)

		fn := func(t db.TransactionNode) (interface{}, error) {
			var currentValue int
			if err := t.Unmarshal(&currentValue); err != nil {
				return currentValue, err
			}
			if currentValue < 0 {
				return 0, nil
			}
			return currentValue + 1, nil
		}

		if err := ref.Transaction(c, fn); err != nil {
			log.Println("Transaction failed to commit:", err)
		}

		err := queue.Dispatcher.Dispatch(func() {
			dataPackage := output.(gin.H)

			firebase.SendSingleNotif(dataPackage["notifToken"].(string), &firebase.Message{
				Title: "New Request!",
				Body:  dataPackage["name"].(string) + " has requested to be your link 🔗",
			})
		})
		if err != nil {
			log.Print(err.Error())
		}

		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  output,
		})
		return
	}
}

func GetOpenReceivedRequests(c *gin.Context) {
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
		log.Println(err.Error())
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
		log.Println(err.Error())
		return
	}

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userA:User)-[link:REQUESTED]->(userB:User {user_id: $user_id})\nRETURN userA.user_id AS id, "+
				"userA.name AS name, userA.bio AS bio, userA.profilepic AS profilepic, link.link_id AS linkId SKIP $offset LIMIT $limit;",
			gin.H{
				"user_id": uid,
				"offset":  offset,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]*models.OpenRequests, 0)
		for data.Next() {
			record := data.Record()
			records = append(records, &models.OpenRequests{
				User: &models.UserBasic{
					UID:        ValueExtractor(record.Get("id")).(string),
					Name:       ValueExtractor(record.Get("name")).(string),
					Bio:        ValueExtractor(record.Get("bio")).(string),
					ProfilePic: ValueExtractor(record.Get("profilepic")).(string),
				},
				LinkId: ValueExtractor(record.Get("linkId")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data.([]*models.OpenRequests),
	})
}

func GetOpenSentRequests(c *gin.Context) {
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
		log.Println(err.Error())
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
		log.Println(err.Error())
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
			"MATCH (userA:User {user_id: $user_id})-[link:REQUESTED]->(userB:User)"+
				"RETURN userA.id AS id, userA.name AS name, userA.bio AS bio, userA.profilepic AS profilepic, link.link_id AS linkId "+
				"SKIP $offset LIMIT $limit;",
			gin.H{
				"user_id": uid,
				"offset":  offset,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]*models.OpenRequests, 0)
		for data.Next() {
			record := data.Record()
			records = append(records, &models.OpenRequests{
				User: &models.UserBasic{
					UID:        ValueExtractor(record.Get("id")).(string),
					Name:       ValueExtractor(record.Get("name")).(string),
					Bio:        ValueExtractor(record.Get("bio")).(string),
					ProfilePic: ValueExtractor(record.Get("profilePic")).(string),
				},
				LinkId: ValueExtractor(record.Get("linkId")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data.([]*models.OpenRequests),
	})
}

func GetFromSocials(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

	permissions, err := getPermissions(uid.(string), c.Param("id"))
	if err != nil {
		switch err.Error() {
		case "no link found":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Link not found",
				"data":  nil,
			})
			log.Println(err.Error())
			return
		case "request exists":
			c.JSON(http.StatusNotFound, gin.H{
				"error": "This relationship is currently a request",
				"data":  nil,
			})
			log.Println(err.Error())
			return
		default:
			c.JSON(http.StatusOK, gin.H{
				"error": nil,
				"data":  nil,
			})
			return
		}
	}

	get, err := firebase.Firestore.Collection("socials").Doc(c.Param("id")).Get(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		log.Println(err.Error())
		return
	}

	socialMedia := get.Data()
	socials := models.Socials{
		Instagram:         "",
		Snapchat:          "",
		Facebook:          "",
		Twitter:           "",
		LinkedIn:          "",
		ProfessionalEmail: "",
		PersonalEmail:     "",
		Tiktok:            "",
		Venmo:             "",
		Website:           "",
		Phone:             "",
	}
	if permissions[11] {
		socials.Website = socialMedia["website"].(string)
	}
	if permissions[10] {
		socials.ProfessionalEmail = socialMedia["professionalEmail"].(string)
	}
	if permissions[9] {
		socials.LinkedIn = socialMedia["linkedin"].(string)
	}
	if permissions[8] {
		socials.Venmo = socialMedia["venmo"].(string)
	}
	if permissions[7] {
		socials.Twitter = socialMedia["twitter"].(string)
	}
	if permissions[6] {
		socials.Tiktok = socialMedia["tiktok"].(string)
	}
	if permissions[5] {
		socials.Facebook = socialMedia["facebook"].(string)
	}
	if permissions[4] {
		socials.Snapchat = socialMedia["snapchat"].(string)
	}
	if permissions[3] {
		socials.Instagram = socialMedia["instagram"].(string)
	}
	if permissions[2] {
		socials.PersonalEmail = socialMedia["personalEmail"].(string)
	}
	if permissions[1] {
		socials.Phone = socialMedia["phone"].(string)
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  socials,
	})
}

func GetToSocials(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Auth mis-match",
			"data":  nil,
		})
		return
	}
	permissions, err := getPermissions(c.Param("id"), uid.(string))

	if err != nil {
		if err.Error() != "no link found" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error: Please Try Again",
				"data":  nil,
			})
			log.Println(err.Error())
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error": nil,
				"data":  -1,
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  permissions,
	})
}

func getPermissions(uidA string, uidB string) (permissions [models.NUM_SOCIALS]bool, e error) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	output, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			`MATCH (userA:User {user_id:$userAId}) MATCH (userB:User {user_id:$userBId}) 
			OPTIONAL MATCH (userA)-[l:LINKED]->(userB) RETURN l.permissions AS linkPermissions,exists((userA)-[:REQUESTED]->(userB)) AS requestExists`,
			gin.H{
				"userAId": uidA,
				"userBId": uidB,
			},
		)
		if err != nil {
			return nil, err
		}

		if record.Next() {
			if ValueExtractor(record.Record().Get("requestExists")).(bool) {
				return nil, errors.New("request exists")
			} else if isNilFixed(ValueExtractor(record.Record().Get("linkPermissions"))) {
				return nil, errors.New("no link found")
			}
			return ValueExtractor(record.Record().Get("linkPermissions")), nil
		} else if record.Err() != nil {
			return nil, record.Err()
		}
		return nil, nil
	})

	if err != nil {
		return permissions, err
	}

	permissionsString := strconv.FormatInt(output.(int64), 2)
	for len(permissionsString) < 12 {
		permissionsString = "0" + permissionsString
	}
	permissionsArr := strings.Split(permissionsString, "")
	for i := 0; i < 12; i++ {
		permissions[i] = permissionsArr[i] == "1"
	}

	return permissions, nil
}

func GetAllLinks(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

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
		log.Println(err.Error())
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
		log.Println(err.Error())
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"CALL {MATCH (userA:User {user_id: $user_id})-[:LINKED]->(userB:User) RETURN userB.name AS name, userB.user_id AS id, userB.profilepic AS profilepic, userB.bio AS bio UNION MATCH (userA:User)-[:LINKED]->(userB:User {user_id: $user_id}) RETURN userA.name AS name, userA.user_id AS id, userA.profilepic AS profilepic, userA.bio AS bio} RETURN name, id, profilepic, bio ORDER BY name SKIP $offset LIMIT $limit",
			gin.H{
				"user_id": uid,
				"offset":  offset,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]*models.UserBasic, 0)
		for data.Next() {
			records = append(records, &models.UserBasic{
				Name:       ValueExtractor(data.Record().Get("name")).(string),
				Bio:        ValueExtractor(data.Record().Get("bio")).(string),
				ProfilePic: ValueExtractor(data.Record().Get("profilepic")).(string),
				UID:        ValueExtractor(data.Record().Get("id")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data.([]*models.UserBasic),
	})
}

func GetLastCheckedInLocations(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
		return
	}

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
		log.Println(err.Error())
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
		log.Println(err.Error())
		return
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id})-[link:LINKED]->(userB:User)-[a:ATTENDED]->(e:Events) WHERE userB.checkedIn <> '' AND link.permissions >= 2048"+
				" AND (e.isPrivate=FALSE OR exists((e)-[:INVITED]->(userA))) RETURN userB.name AS name, userB.user_id AS id, userB.profilepic AS profilepic,"+
				"e.name AS eventName, e.event_id AS eventId, e.type AS eventType ORDER BY a.timeAttended DESC SKIP $offset LIMIT $limit",
			gin.H{
				"user_id": uid,
				"offset":  offset,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}
		records := make([]*models.LastCheckInLocation, 0)
		for data.Next() {
			records = append(records, &models.LastCheckInLocation{
				User: &models.UserBasic{
					Name:       ValueExtractor(data.Record().Get("name")).(string),
					UID:        ValueExtractor(data.Record().Get("id")).(string),
					ProfilePic: ValueExtractor(data.Record().Get("profilepic")).(string),
				},
				EventName: ValueExtractor(data.Record().Get("eventName")).(string),
				EventID:   ValueExtractor(data.Record().Get("eventId")).(string),
				EventType: ValueExtractor(data.Record().Get("eventType")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data,
	})
}

func UpdatePermissions(c *gin.Context) {
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

	var jsonData models.Link
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(), //TODO log marshall error
			"data":  nil,
		})
		return
	}

	jsonData.Me = &models.UserBasic{
		UID: uid.(string),
	}

	jsonData.UserRec = &models.UserBasic{
		UID: c.Param("id"),
	}

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {user_id: $user_rec.uid})-[link:LINKED]->(:User {user_id: $me.uid}) "+
				"SET link.permissions = $permissions;",
			structToDbMap(jsonData),
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
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Successfully Updated!",
	})
}
