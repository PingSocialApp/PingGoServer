package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"net/http"
	dbclient "pingserver/src/db_client"
)

func GetUserBasic(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	//TODO add rest of query
	transaction, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) RETURN userA.name",
			map[string]interface{}{
				"uid": c.Param("uid"),
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
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"name": transaction,
		})
	}
}

func GetUserSocials(c *gin.Context) {

}

func CheckoutUser(c *gin.Context) {

}

func CreateNewUser(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	input := map[string]interface{}{
		"uid": c.Param("uid"),
		"name": c.PostForm("name"),
		"number": c.PostForm("number"),
	}

	//TODO add rest of query
	transaction, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MERGE (userA:User {user_id:$uid, name:$name})-[:DIGITAL_PROFILE]->" +
				"(social:Social{number:$number})",
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
			"error": transaction.(string),
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"message": c.PostForm("name") + " has been created",
		})
	}
}

func UpdateUserInfo(c *gin.Context) {

}

func SetNotifToken(c *gin.Context) {
	data := dbclient.CreateSession()
	defer data.Close()

	_, err := data.WriteTransaction(func(transaction neo4j.Transaction) (interface {}, error){
		result, err := transaction.Run(
			"MATCH (userA:User {user_id: $uid}) SET user.notifToken = $token",
			map[string]interface{}{
				"uid": c.Param("uid"),
				"token": c.PostForm("token"),
			})
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
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"message": "Notification token updated",
		})
	}
}
