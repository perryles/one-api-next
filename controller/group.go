package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/perryles/one-api-next/model"
)

func GetGroups(c *gin.Context) {
	groupNames := model.GetGroupNames()
	if len(groupNames) == 0 {
		groupNames = []string{"default"}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}