package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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

	//just want a flow of events for some debugging
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	cleanup := func() {
		log.Println("close the websocket conn")
		conn.Close()
		//
		log.Println("stop the ticker")
		ticker.Stop()
	}
	defer cleanup()

	spamcount := 0

	for {
		select {
		case message := <-player.MessageChannel:
			log.Printf("sending a message like this: %+v", message)
			if err := conn.WriteJSON(message); err != nil {
				// Handle error: failed to send message
				log.Println("Failed to send", message)
				log.Println("Write error:", err)
				return
			}
		case <-ticker.C:
			spamcount++
			// Send a message every tick
			//messageData := struct {
			//	MessageType string `json:"messageType"`
			//	Content     string `json:"content"`
			//	Count       int    `json:"count"`
			//}{
			//	MessageType: "Spam",
			//	Content:     fmt.Sprintf("Spam message from server"),
			//	Count:       spamcount,
			//}
			//
			//// Marshal the struct to JSON
			//messageJSON, err := json.Marshal(messageData)
			//if err != nil {
			//	log.Println("JSON marshal error:", err)
			//	return // Exit the loop and end the connection on error
			//}
			//
			//// Send the JSON message
			//if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			//	log.Println("Write error:", err)
			//	return // Exit the loop and end the connection on error
			//}
		}

	}

}
