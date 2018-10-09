package main

import (
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	eztemplate "github.com/michelloworld/ez-gin-template"
	"github.com/prgra/monik/abills"
	"github.com/stvp/go-toml-config"
)

var myurl = config.String("abills.url", "")

func main() {

	err := config.Parse("./config.toml")
	if err != nil {
		panic(err)
	}
	log.Print(*myurl)
	err = abills.New(*myurl)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	render := eztemplate.New()
	render.TemplatesDir = "views/" // default
	render.Layout = "layouts/base" // default
	render.Ext = ".html"           // default
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	render.TemplateFuncMap = template.FuncMap(funcMap)
	r.Use(gin.Recovery())
	r.Use(FaviconNew("./static/favicon.ico"))

	r.GET("/", offline)
	r.GET("/search", search)
	r.GET("/search/:q", search)
	r.GET("/hist/:id", history)
	r.HTMLRender = render.Init()
	r.Static("/static", "./static")

	r.Run(":3001")
}
