package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	dbclient "pingserver/db_client"
	firebase "pingserver/firebase_client"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//Requests
func DeleteRequest(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (:User {uid: $uid})-[link:REQUESTED {link_id:$lid}]->(:User)\n DELETE link;",
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
			"error": err.Error(),
			"data":  nil,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Request Deleted",
	})

}

func AcceptRequest(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (userA:User {})-[request:REQUESTED {link_id: $link_id}]->(userB:User)"+
				"MERGE (userA:User)-[link:LINKED {link_id: request.link_id, permissions: request.permissions}]->(userB:User)"+
				"DELETE request",
			gin.H{
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
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  "Request Accepted",
	})
}

func DeclineRequest(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, "http://localhost:8080/"+c.Param("rid")+"/delete")
}

func SendRequest(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	var jsonData gin.H // map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if e := json.Unmarshal(data, &jsonData); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
			"data":  nil,
		})
		return
	}

	jsonData["userA_id"] = uid

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	output, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		exists, err := transaction.Run(
			"MATCH (userA:User {user_id: $userA_id}) MATCH (userB:User {user_id: $userB_id})"+
				"RETURN EXISTS (userA)-[:REQUESTED]->(userB) OR (userA)-[:LINKED]->(userB) AS linkExists",
			jsonData,
		)
		if err != nil {
			return nil, err
		}

		if exists.Record().GetByIndex(0).(bool) {
			return "exists", nil
		}

		data, err := transaction.Run(
			"MATCH (userA:User {user_id: $userA_id}) MATCH (userB:User {user_id: $userB_id})"+
				"MERGE (userA)-[r:REQUESTED {link_id: apoc.create.uuid(), permissions: $perm}]->(userB) RETURN r.link_id AS linkId",
			jsonData,
		)

		if err != nil {
			return nil, err
		}

		if data.Next() {
			return gin.H{
				"id": ValueExtractor(data.Record().Get("linkId")),
			}, nil
		}
		return nil, data.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
	}

	switch output {
	case "exists":
		c.JSON(http.StatusConflict, gin.H{
			"error": nil,
			"data":  "Record already exists",
		})
	case "created":
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"data": gin.H{
				"id": jsonData["lid"],
			},
		})
	}
}

func GetOpenSentRequests(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
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
		records := make([]interface{}, 0)
		for data.Next() {
			records = append(records, gin.H{
				"id":         ValueExtractor(data.Record().Get("id")).(string),
				"name":       ValueExtractor(data.Record().Get("name")).(string),
				"bio":        ValueExtractor(data.Record().Get("bio")).(string),
				"profilepic": ValueExtractor(data.Record().Get("profilepic")).(string),
				"link_id":    ValueExtractor(data.Record().Get("linkId")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data,
	})
}

func GetOpenReceivedRequests(c *gin.Context) {
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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
		records := make([]interface{}, 0)
		for data.Next() {
			records = append(records, gin.H{
				"id":         ValueExtractor(data.Record().Get("id")).(string),
				"name":       ValueExtractor(data.Record().Get("name")).(string),
				"bio":        ValueExtractor(data.Record().Get("bio")).(string),
				"profilepic": ValueExtractor(data.Record().Get("profilepic")).(string),
				"link_id":    ValueExtractor(data.Record().Get("linkId")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data.([]interface{}),
	})
}

//Links
func GetFromSocials(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	permissions, err := getPermissions(c.Param("id"), uid.(string))
	if err != nil {
		if err.Error() != "No Link Found" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"data":  nil,
			})
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
	socials := gin.H{
		"instagram":         nil,
		"snapchat":          nil,
		"facebook":          nil,
		"twitter":           nil,
		"linkedin":          nil,
		"professionalEmail": nil,
		"personalEmail":     nil,
		"tiktok":            nil,
		"venmo":             nil,
		"website":           nil,
		"phone":             nil,
		"location":          permissions[0],
	}
	if permissions[11] {
		socials["website"] = socialMedia["website"]
	}
	if permissions[10] {
		socials["professionalEmail"] = socialMedia["professionalEmail"]
	}
	if permissions[9] {
		socials["linkedin"] = socialMedia["linkedin"]
	}
	if permissions[8] {
		socials["venmo"] = socialMedia["venmo"]
	}
	if permissions[7] {
		socials["twitter"] = socialMedia["twitter"]
	}
	if permissions[6] {
		socials["tiktok"] = socialMedia["tiktok"]
	}
	if permissions[5] {
		socials["facebook"] = socialMedia["facebook"]
	}
	if permissions[4] {
		socials["snapchat"] = socialMedia["snapchat"]
	}
	if permissions[3] {
		socials["instagram"] = socialMedia["instagram"]
	}
	if permissions[2] {
		socials["personalEmail"] = socialMedia["personalEmail"]
	}
	if permissions[1] {
		socials["phone"] = socialMedia["phone"]
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
				"error": err.Error(),
				"data":  nil,
			})
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
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
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
		records := make([]interface{}, 0)
		for data.Next() {
			records = append(records, gin.H{
				"name":       ValueExtractor(data.Record().Get("name")).(string),
				"bio":        ValueExtractor(data.Record().Get("bio")).(string),
				"profilepic": ValueExtractor(data.Record().Get("profilepic")).(string),
				"id":         ValueExtractor(data.Record().Get("id")).(string),
			})
		}
		return records, data.Err()
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
		"data":  data,
	})
}

func GetLastCheckedInLocations(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID not set from Authentication",
			"data":  nil,
		})
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
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
		records := make([]interface{}, 0)
		for data.Next() {
			records = append(records, gin.H{
				"name":       ValueExtractor(data.Record().Get("name")).(string),
				"id":         ValueExtractor(data.Record().Get("id")).(string),
				"profilepic": ValueExtractor(data.Record().Get("profilepic")).(string),
				"eventName":  ValueExtractor(data.Record().Get("eventName")).(string),
				"eventId":    ValueExtractor(data.Record().Get("eventId")).(string),
				"eventType":  ValueExtractor(data.Record().Get("eventType")).(string),
			})
		}
		return records, data.Err()
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
	}

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

	jsonData["myUID"] = uid
	jsonData["uidB"] = c.Param("id")

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(
			"MATCH (userA:User {user_id: $myUID})-[link:LINKED]->(userB:User{user_id: $uidB}) "+
				"SET link.permissions = $permissions;",
			jsonData,
		)

		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  nil,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data": gin.H{
			"code": jsonData["permissions"],
		},
	})
}
