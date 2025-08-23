package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Define an event struct to hold the data we'll send to clients.
type Event struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

var eventChannel = make(chan Event)

func main() {
	gin.SetMode(gin.DebugMode)
	// Use gin.New() for a more minimal setup without default middleware.
	router := gin.Default()

	// Register the long-polling handler.
	router.GET("/poll", longPollHandler)

	// Register the event publisher handler. This is a simple endpoint to
	// simulate a new event being created.
	router.POST("/publish", publishHandler)

	// Start a goroutine to continuously publish new events to the channel.
	// In a real application, this would be triggered by some internal logic.
	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second) // Publish a new event every 5 seconds.
	// 		eventChannel <- Event{
	// 			Message: "New message from the server! 🚀",
	// 			Time:    time.Now(),
	// 		}
	// 		// log.Printf("Published a new event at %s", time.Now().Format("2006-01-02 15:04:05"))
	// 	}
	// }()

	log.Println("Server started on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// longPollHandler handles the long-polling requests.
func longPollHandler(c *gin.Context) {
	// Set a timeout for the request to prevent it from hanging forever.
	// This is a crucial step in real-world long polling implementations.
	timeout := time.After(30 * time.Second)

	select {
	case event := <-eventChannel:
		// A new event was received. Respond to the client immediately.
		c.JSON(http.StatusOK, event)
		return
	case <-timeout:
		// The timeout was reached. Respond with a "no content" status.
		c.JSON(http.StatusNoContent, gin.H{"message": "No new events."})
		return
	}
}

// publishHandler simulates publishing a new event.
func publishHandler(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventChannel <- Event{
		Message: req.Message,
		Time:    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event published."})
}
