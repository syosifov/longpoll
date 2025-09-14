package main

import (
	"log"

	"github.com/syosifov/longpoll/lpoll"

	"github.com/gin-gonic/gin"
)

// Event is the data structure for events.

func main() {
	// log.Printf("Client timeout set to %s", clientTimeout.String())

	router := gin.Default()

	// Handler for clients to subscribe and start polling.
	// router.GET("/subscribe/:clientId", subscribeHandler)

	// Handler for long-polling requests.
	router.GET("/poll/:clientId", pollHandler)

	// Handler to publish events to a specific client.
	router.POST("/publish/:clientId", publishHandler)

	// Clean up inactive clients periodically.

	log.Println("Server started on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// Use wrapper functions to call the handlers in lpoll package
func pollHandler(c *gin.Context) {
	lpoll.PollHandler(c)
}

func publishHandler(c *gin.Context) {
	lpoll.PublishHandler(c)
}
