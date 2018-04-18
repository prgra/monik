package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func offline(c *gin.Context) {
	c.HTML(http.StatusOK, "main/offline", nil)
}
