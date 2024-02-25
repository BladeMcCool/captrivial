package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (gs *GameServer) GameStatusHandler(c *gin.Context) {
	lobbyId := c.Param("lobbyId")
	if lobbyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lobby ID"})
		return
	}

	// Assuming you have a method GetLobby that retrieves a lobby by its ID
	lobby, exists := gs.Lobbies.GetLobby(lobbyId)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lobby ID"})
		return
	}

	// Get the game status from the lobby
	status := lobby.GameStatus()

	// Directly pass the GameStatusResult struct to c.JSON
	c.JSON(http.StatusOK, status)
}
