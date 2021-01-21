package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"pingserver/db_client"
)

// GetUserBasic Get user's name, profile pic, and bio
func GetUserBasic(c *gin.Context) {

}

func GetUserSocials(c *gin.Context) {

}

func CheckoutUser(c *gin.Context) {

}

func CreateNewUser(c *gin.Context) {
	session := dbclient.CreateSession()
	defer dbclient.KillSession(session)

	data := map[string]interface{}{
		"uid":        c.Param("id"),
		"name":       c.Query("name"),
		"bio":        c.Query("bio"),
		//"profilepic": c.Query("profilepic"),
		//"sex":        c.Query("sex"),
		"number":     c.Query("number"),
		//"perEmail":   c.Query("perEmail"),
		//"ig":         c.Query("ig"),
		//"sc":         c.Query("sc"),
		//"fb":         c.Query("fb"),
		//"tt":         c.Query("tt"),
		//"tw":         c.Query("tw"),
		//"proEmail":   c.Query("proEmail"),
		//"li":         c.Query("li"),
		//"website":    c.Query("website"),
		//"venmo":      c.Query("venmo"),
	}

	result, err := session.Run(
		"CREATE (u:User {"+
			"uid: $uid" + "name: $name"+ "bio: $bio})," +
			//"profilepic: $profilepic"+ "sex: $sex}), "+
			"(s:Socials{" +
			"number: $number"+
			//"perEmail: $perEmail"+
			//"ig: $ig"+
			//"sc: $sc"+
			//"fb: $fb"+
			//"tt: $tt"+
			//"tw: $tw"+
			//"proEmail: $proEmail"+
			//"li: $li"+
			//"website: $website"+
			//"venmo: $venmo}"+
			"}), (u)-[r:DIGITAL_PROFILE]=>(s)", data)

	_ = result

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		log.Print("name:" + c.Query("name"))
		c.JSON(http.StatusAccepted, gin.H{
			"error": "NaN",
			"message": "User" + c.Query("id") + "("+ c.Query("name")+") has been made",
		})
	}
}

func UpdateUserInfo(c *gin.Context) {

}

func SetNotifToken(c *gin.Context) {

}
