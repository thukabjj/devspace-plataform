package main

import (
	"devspace-backend/handler"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// define routes
	r.POST("/createRepository", handler.CreateRepositoryHandler)

	// run server
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err.Error())
	}
}
