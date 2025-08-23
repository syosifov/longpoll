package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Event is the data structure for events.
type Event struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// ClientState holds the channel and a timestamp for a specific client.
type ClientState struct {
	Channel  chan Event
	LastSeen time.Time
}

var (
	// clientChannels maps a client ID to its state.
	clientChannels = make(map[string]*ClientState)
	// mutex for safe concurrent access to the clientChannels map.
	mu sync.RWMutex
	// Define the timeout duration for client inactivity.
	clientTimeout = 1 * time.Minute
)

func main() {
	log.Printf("Client timeout set to %s", clientTimeout.String())
	router := gin.Default()

	// Handler for clients to subscribe and start polling.
	router.GET("/subscribe/:clientId", subscribeHandler)

	// Handler for long-polling requests.
	router.GET("/poll/:clientId", pollHandler)

	// Handler to publish events to a specific client.
	router.POST("/publish/:clientId", publishHandler)

	// Clean up inactive clients periodically.
	go cleanUpInactiveClients()

	log.Println("Server started on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// subscribeHandler initializes a new channel and sets the initial "last seen" time.
func subscribeHandler(c *gin.Context) {
	// Generate a unique ID for the new client.
	// clientId := uuid.New().String()
	clientId := c.Param("clientId")

	// Create a buffered channel for this client.
	clientChan := make(chan Event, 1)

	// Add the new client and its channel to the global map.
	mu.Lock()
	clientChannels[clientId] = &ClientState{
		Channel:  clientChan,
		LastSeen: time.Now(),
	}
	mu.Unlock()

	log.Printf("Client subscribed: %s", clientId)
	c.JSON(http.StatusOK, gin.H{"clientId": clientId})
}

// pollHandler handles the long-polling requests for a specific client.
func pollHandler(c *gin.Context) {
	clientId := c.Param("clientId")
	if clientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clientId is required"})
		return
	}

	mu.Lock()
	client, ok := clientChannels[clientId]
	if !ok {
		clientChan := make(chan Event, 1)
		client = &ClientState{
			Channel:  clientChan,
			LastSeen: time.Now(),
		}
		clientChannels[clientId] = client
		log.Printf("Client subscribed: %s", clientId)
	} else {
		client.LastSeen = time.Now()
	}
	mu.Unlock()

	// Set a timeout for the request to prevent it from hanging indefinitely.
	timeout := time.After(30 * time.Second)

	select {
	case event := <-client.Channel:
		// An event was received. Respond to the client.
		c.JSON(http.StatusOK, event)
		return
	case <-timeout:
		// The timeout was reached. Respond with "no content."
		c.JSON(http.StatusNoContent, nil)
		log.Printf("Poll timeout for client: %s", clientId)
		return
	}
}

// publishHandler publishes an event to a specific client identified by its ID.
func publishHandler(c *gin.Context) {
	clientId := c.Param("clientId")
	if clientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clientId is required"})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.RLock()
	client, ok := clientChannels[clientId]
	mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	select {
	case client.Channel <- Event{Message: req.Message, Time: time.Now()}:
		c.JSON(http.StatusOK, gin.H{"message": "Event published."})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Client channel is full, skipping event."})
	}
}

// cleanUpInactiveClients periodically removes clients that haven't re-polled.
func cleanUpInactiveClients() {
	for {
		time.Sleep(1 * time.Minute) // Check for inactive clients every minute.

		mu.Lock()
		for clientId, clientState := range clientChannels {
			// Check if the client's last seen time is older than the timeout.
			if time.Since(clientState.LastSeen) > clientTimeout {
				delete(clientChannels, clientId)
				log.Printf("Cleaned up inactive client: %s", clientId)
				log.Printf("Active clients remaining: %d", len(clientChannels))
				// Close the channel to release resources.
				close(clientState.Channel)
			}
		}
		mu.Unlock()
	}
}
