package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"io"
	"net/http"
	"pingserver/broker"
	dbclient "pingserver/db_client"
)

var markersManager = broker.NewRoomManager()
var eventManager = broker.NewRoomManager()

func GetUserBasic(c *gin.Context) {
	data := dbclient.CreateSession()
	defer dbclient.KillSession(data)

	//TODO add rest of query
	_, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) RETURN userA.name, userA.bio, userA.profilepic",
			map[string]interface{}{
				"uid": "users/" + c.Param("uid"),
			})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values(), nil
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"isEmpty": true,
			"data": nil,
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"isEmpty": false,
			"data": gin.H{
				"name": transaction.([]interface{})[0],
				"bio": transaction.([]interface{})[1],
				"profilepic": transaction.([]interface{})[2],
			},
		})
	}
}

func GetUserSocials(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	//TODO add rest of query
	transaction, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $user_id})-[:DIGITAL_PROFILE]->(social:Socials)\n" +
				"RETURN social.number, social.perEmail, social.ig, social.sc, social.fb, social.tt, " +
				"social.tw, social.venmo, social.proEmail, social.li, social.website",
			map[string]interface{}{
				"uid": "users/" + c.Param("uid"),
			})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values(), nil
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"isEmpty": true,
			"data": nil,
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"isEmpty": false,
			"data": gin.H{
				"number": transaction.([]interface{})[0],
				"perEmail": transaction.([]interface{})[1],
				"ig": transaction.([]interface{})[2],
				"sc": transaction.([]interface{})[3],
				"fb": transaction.([]interface{})[4],
				"tt": transaction.([]interface{})[5],
				"tw": transaction.([]interface{})[6],
				"venmo": transaction.([]interface{})[7],
				"proEmail": transaction.([]interface{})[8],
				"li": transaction.([]interface{})[9],
				"website": transaction.([]interface{})[10],
			},
		})
	}
}

func CheckinUser(c *gin.Context){
	eventManager.Submit(c.PostForm("eventId"), "attendee-added")
}

func CheckoutUser(c *gin.Context) {
	eventManager.Submit(c.PostForm("eventId"), "attendee-removed")
}

func CreateNewUser(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	input := map[string]interface{}{
		// User
		"uid": "users/" + c.Param("uid"),
		"name": c.PostForm("name"),
		"bio": c.PostForm("bio"),
		"profilepic": c.PostForm("profilepic"),
		"userType": c.PostForm("userType"),
		"rating": 3.5,

		"number": c.PostForm("number"),
		"perEmail": c.PostForm("perEmail"),
		"ig": c.PostForm("ig"),
		"sc": c.PostForm("sc"),
		"fb": c.PostForm("fb"),
		"tt": c.PostForm("tt"),
		"tw": c.PostForm("tw"),
		"venmo": c.PostForm("venmo"),
		"proEmail": c.PostForm("proEmail"),
		"li": c.PostForm("li"),
		"website": c.PostForm("website"),
	}

	//TODO add rest of query
	_, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MERGE (userA:User {user_id:$uid, name:$name, bio:$bio, profilepic:$profilepic, " +
				"userType:$userType, rating:$rating})-[:DIGITAL_PROFILE]->" +
				"(social:Socials{number:$number, perEmail:$perEmail, ig:$ig, sc:$sc, fb:$fb, tt:$tt, tw:$tw, venmo:$venmo," +
				"proEmail:$proEmail, li:$li, website:$website})",
		input)
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
			"message": err.Error(),
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"message": c.PostForm("name") + " has been created",
		})
	}
}

func UpdateUserInfo(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	//TODO add rest of query
	transaction, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: @}) \nSET userA.user_id=@, userA.name =@,  userA.bio=@," +
				" userA.profilepic=@, userA.sex=@, userA.notifToken=@, userA.userType=@, userA.rating=@ ",
			map[string]interface{}{
				"uid": "users/" + c.Param("uid"),
			})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values(), nil
		}

		return nil, result.Err()
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"isEmpty": true,
			"data": nil,
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"isEmpty": false,
			"data": gin.H{
				"name": transaction.([]interface{})[0],
				"bio": transaction.([]interface{})[1],
				"profilepic": transaction.([]interface{})[2],
			},
		})
	}
}

func GetRelevantMarkers(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	roomid := c.Param("uid")
	listener := markersManager.OpenListener(roomid)
	defer markersManager.CloseListener(roomid, listener)
	defer markersManager.DeleteBroadcast(roomid)

	clientGone := c.Writer.CloseNotify()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case message := <-listener:
			c.SSEvent("message", message)
			return true
		}
	})

}
