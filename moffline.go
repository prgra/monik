package main

import (
	"log"
	"net/http"
	"sort"

	"github.com/prgra/monik/abills"

	"github.com/gin-gonic/gin"
)

func offline(c *gin.Context) {
	off := abills.GetOffline()
	// tst := abills.NasGrep(off, "34:08:04:5e:67:50")
	var s abills.SortedNases
	s = off
	sort.Sort(s)
	cnt := len(s)
	c.HTML(http.StatusOK, "main/offline", gin.H{"offline": s, "cnt": cnt, "q": ""})
}

func search(c *gin.Context) {
	log.Print("log2")
	all := abills.GetAllNases()
	q := c.Param("q")
	if q == "" {
		q = c.Query("q")
	}
	rst := abills.NasGrep(all, q)
	var s abills.SortedNases
	s = rst
	sort.Sort(s)
	cnt := len(s)
	c.HTML(http.StatusOK, "main/offline", gin.H{"offline": s, "cnt": cnt, "q": q})
}
