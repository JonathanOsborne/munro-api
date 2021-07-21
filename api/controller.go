package api

import (
	"context"

	"github.com/JonathanOsborne/munro-api/db"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	Munros      map[string]Munro
	Router      *gin.Engine
	MongoClient *db.DbClient
}

func NewController() (con *Controller) {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	cli := db.NewDBClient(context.Background())

	con = &Controller{
		Router:      r,
		MongoClient: cli,
	}

	con.SetMunrosData()
	con.MountEndpoints()

	return con
}

func (con *Controller) MountEndpoints() {
	r := con.Router

	r.GET("/munro/:name", con.GetMunro)
	r.GET("/munro", con.ListMunroNames)
	r.GET("/munros", con.ListMunros)

}
