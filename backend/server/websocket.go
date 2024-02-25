package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

func (gs *GameServer) WsHandler(c *gin.Context) {
	w := c.Writer
	r := c.Request

	// Extracting the lobby ID and player sessionId from the URL path
	lobbyId := c.Param("lobbyId")
	if lobbyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lobby ID"})
		return
	}
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID"})
		return
	}

	lobby, found := gs.Lobbies.GetLobby(lobbyId)
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lobby doesnt exist"})
		return
	}
	player, _ := lobby.GetPlayer(sessionId) //TODO decide and make consistent the return style of these getters. idk probably dont actually need to send and error
	if player == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player not found"})
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}

	cleanup := func() {
		log.Println("close the websocket conn")
		conn.Close()
	}
	defer cleanup()

	for {
		select {
		case message, ok := <-player.MessageChannel:
			if !ok {
				// The channel was closed; exit the loop
				log.Println("Message channel closed.")
				return
			}
			log.Printf("sending a message like this: %+v", message)
			if err := conn.WriteJSON(message); err != nil {
				// Handle error: failed to send message
				log.Println("Failed to send", message)
				log.Println("Write error:", err)
				return
			}
		}
	}
}
