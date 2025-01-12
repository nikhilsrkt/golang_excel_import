package main

import (
	"excel_project/dialects"
	"excel_project/routers"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.Default()
	if _, err := dialects.GetConnection(); err != nil {
		log.Panic(fmt.Printf("error connectin to DB: %s", err))
	}
	redis := dialects.RedisClient
	go redis.Connect()
	routers.Endpoints(app)
	app.Run(":8080")
}
