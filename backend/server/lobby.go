package server

import (
	"github.com/ProlificLabs/captrivia/game"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (gs *GameServer) NewLobbyHandler(c *gin.Context) {
	var gameParams struct {
		QuestionCount int `json:"questionCount"`
		CountdownMs   int `json:"countdownMs"`
	}
	if err := c.ShouldBindJSON(&gameParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	sessionID := gs.generateSessionID()
	lobbyID := gs.Lobbies.AddLobby(gameParams.QuestionCount, gameParams.CountdownMs, &game.Player{
		SessionID:         sessionID,
		Score:             0,
		QuestionsAnswered: []string{},
	})
	c.JSON(http.StatusOK, gin.H{"sessionId": sessionID, "lobbyId": lobbyID, "questionCount": gameParams.QuestionCount, "countdownMs": gameParams.CountdownMs})
}

func (gs *GameServer) JoinLobbyHandler(c *gin.Context) {
	// Extracting the lobby ID from the URL path
	lobbyId := c.Param("lobbyId")
	if lobbyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lobby ID"})
		return
	}

	// Assuming you have a method to retrieve a lobby by its ID and another to add a player to a lobby
	lobby, found := gs.Lobbies.GetLobby(lobbyId)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lobby not found"})
		return
	}

	// AddPlayer is a method that adds a player to the specified lobby and returns an error if it fails
	log.Printf("JoinLobbyHandler, put another player into lobby id: %s", lobbyId)
	sessionId := gs.generateSessionID() //treat as a new player when joining a lobby. session is to identify the player within the lobby.
	err := lobby.AddPlayer(sessionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join lobby: " + err.Error()})
		return
	}

	// Respond with a success message or other relevant information
	c.JSON(http.StatusOK, gin.H{"message": "Joined lobby successfully", "lobbyId": lobbyId, "sessionId": sessionId})
}
