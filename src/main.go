package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kisbogdan-kolos/gallery/api"
	"github.com/kisbogdan-kolos/gallery/db"
	"github.com/kisbogdan-kolos/gallery/helper"
)

func main() {
	err := db.DbConnect()

	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	api.Register(router.Group("/api"))

	addr := helper.EnvGet("SERVER_ADDRESS", "0.0.0.0:8080")

	router.Run(addr)
}
