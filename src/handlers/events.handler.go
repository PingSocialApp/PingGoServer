package handlers

import (
	"github.com/gin-gonic/gin"
	"io"
)

func DeleteEvent(c *gin.Context) {
}

func GetAttendees(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	roomid := c.Param("id")
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

func GetEventDetails(c *gin.Context) {
}

func GetInPartyDetails(c *gin.Context) {

}

func UpdateEvent(c *gin.Context) {

}

func CreateEvent(c *gin.Context) {
}
