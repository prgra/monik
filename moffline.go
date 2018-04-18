package main

import (
	"github.com/gin-gonic/gin"
)

func offline(g *gin.Context) {
	g.String(200, "ok")
}
