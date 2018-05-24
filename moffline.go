package main

import (
	"monik/abills"
	"net/http"

	"github.com/gin-gonic/gin"
)

func offline(c *gin.Context) {
	offs, on := abills.GetOffline()
	
	c.HTML(http.StatusOK, "main/offline", gin.H{"offline": offs, "on": on})
}
