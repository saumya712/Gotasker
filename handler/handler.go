package handler

import (
	"practice/domain"
	"practice/logic"
	"strconv"

	"github.com/gin-gonic/gin"
)

type requestmsg struct {
	Tasks []domain.Task `json:"tasks"`
}

func Posttask(c *gin.Context) {
	var reqmsg requestmsg

	if err := c.ShouldBindJSON(&reqmsg); err != nil {
		c.JSON(401, gin.H{
			"error": "check the input type",
		})
		return
	}

	logic.Store.Addtasks(reqmsg.Tasks)

	go logic.Processtasks(reqmsg.Tasks)

	c.JSON(200, gin.H{"msg": "task processed successfully😊"})

}

func Getthetask(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)

	if err != nil {
		c.JSON(401, gin.H{"error": "error"})
		return
	}
	result, ok := logic.Store.Gettask(id)
	if !ok {
		c.JSON(401, gin.H{"error": "error"})
		return
	}

	c.JSON(200, result)
}
