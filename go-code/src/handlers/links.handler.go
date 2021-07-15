package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"
	firebase "pingserver/firebase_client"
	"pingserver/models"
	"strconv"
	"strings"

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
			"MATCH (:User {uid: $uid})-[link:REQUESTED {link_id:$lid}]->(:User) DETACH DELETE link;",
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
		fmt.Println(err.Error())
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

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (userA:User)-[request:REQUESTED {link_id: $link_id}]->(userB:User {user_id: $uid})"+
				" MERGE (userA:User)-[link:LINKED {link_id: request.link_id, permissions: request.permissions}]->(userB:User)"+
				" DETACH DELETE request",
			gin.H{
				"uid":     uid,
				"link_id": c.Param("rid"),
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
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Request Accepted",
	})
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
			"MATCH (:User)-[link:REQUESTED {link_id:$lid}]->(:User {uid: $uid}) DETACH DELETE link;",
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
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Request Deleted",
	})
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

	var jsonData models.Request // map[string]interface{}
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

	jsonData.UID = uid.(string)

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	output, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		inputs := structToDbMap(jsonData)

		exists, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) MATCH (userB:User {user_id: $user_rec})"+
				" RETURN EXISTS (userA)-[:REQUESTED]->(userB) OR (userA)-[:LINKED]->(userB) AS linkExists",
			inputs,
		)
		if err != nil {
			return nil, err
		}

		if exists.Record().Values[0].(bool) {
			return "exists", nil
		}

		data, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) MATCH (userB:User {user_id: $user_rec})"+
				"MERGE (userA)-[r:REQUESTED {link_id: apoc.create.uuid(), permissions: $permissions}]->(userB) RETURN r.link_id AS linkId",
			inputs,
		)

		if err != nil {
			return nil, err
		}

		if data.Next() {
			return ValueExtractor(data.Record().Get("linkId")).(string), nil
		}
		return nil, data.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error: Please Try Again",
			"data":  nil,
		})
		fmt.Println(err.Error())
		return
	}

	switch output {
	case "exists":
		c.JSON(http.StatusConflict, gin.H{
			"error": nil,
			"data":  "Record already exists",
		})
		return
	case "created":
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data":  output,
		})
		return
	}
}

func GetOpenSentRequests(c *gin.Context) {
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

	data, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userA:User)-[link:REQUESTED]->(userB:User {user_id: $user_id})\nRETURN userA.id AS id, "+
				"userA.name AS name, userA.bio AS bio, userA.profilepic AS profilepic link.link_id AS linkId SKIP $offset LIMIT $limit;",
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
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  data.([]*models.OpenRequests),
	})
}

func GetOpenReceivedRequests(c *gin.Context) {
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
		fmt.Println(err.Error())
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

	permissions, err := getPermissions(c.Param("id"), uid.(string))
	if err != nil {
		if err.Error() != "No Link Found" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error: Please Try Again",
				"data":  nil,
			})
			fmt.Println(err.Error())
			return
		} else {
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
	uid, exists := c.Get("UID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Auth mis-match",
			"data":  nil,
		})
		return
	}

	permissions, err := getPermissions(uid.(string), c.Param("id"))
	if err != nil {
		if err.Error() != "No Link Found" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error: Please Try Again",
				"data":  nil,
			})
			fmt.Println(err.Error())
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

func getPermissions(uidA string, uidB string) (permissions [12]bool, e error) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	output, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		record, err := transaction.Run(
			"MATCH (userA:User {user_id: $userAId})-[link:LINKED]->(userB:User{user_id: $userBid}) RETURN link.permissions;",
			gin.H{
				"userAId": uidA,
				"userBId": uidB,
			},
		)
		if err != nil {
			return nil, err
		}

		if record.Next() {
			return ValueExtractor(record.Record().Get("link.permissions")), nil
		} else if record.Err() != nil {
			return nil, record.Err()
		} else {
			return nil, errors.New("No link found")
		}
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
		fmt.Println(err.Error())
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

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		data, err := transaction.Run(
			"MATCH (userB:User {isCheckedIn:true})-[link:LINKED]->(userA:User {user_id: $user_id})\nWHERE link.permissions >= 2048\n"+
				"MATCH (userB)-[a:ATTENDING]->(e:Events)\nRETURN userB.name AS name, userB.user_id AS id, userB.profilepic AS profilepic,\n"+
				"e.name AS eventName, e.event_id AS eventId, e.type AS eventType ORDER BY DESC a.timeAttended SKIP $offset LIMIT $limit",
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
		fmt.Println(err.Error())
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

	var jsonData models.Link // map[string]interface{}
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

	jsonData.UID = uid.(string)
	jsonData.UserRecUID = c.Param("id")

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {user_id: $myUID})-[link:LINKED]->(:User{user_id: $user_rec_id}) "+
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
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Successfully Updated!",
	})
}
