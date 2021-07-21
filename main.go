package main

import (
	"github.com/JonathanOsborne/munro-api/api"
)

func main() {
	c := api.NewController()
	c.Router.Run(":8080")
}
