package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/davecgh/go-spew/spew"
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

func history(c *gin.Context) {
	sid := c.Param("id")
	id, _ := strconv.Atoi(sid)
	if id == 0 {
		c.String(http.StatusOK, "need id")
	}
	n, ok := abills.GetNas(id)
	if ok {
		c.HTML(http.StatusOK, "main/history", gin.H{"n": spew.Sdump(n)})
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("no nas with id %d", id))
}
