package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (gs *GameServer) StartGameHandler(c *gin.Context) {
	var lobbyParams struct {
		LobbyId   string `json:"lobbyId"`
		SessionId string `json:"sessionId"`
	}
	if err := c.ShouldBindJSON(&lobbyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	lobby, found := gs.Lobbies.GetLobby(lobbyParams.LobbyId)
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find lobby: " + lobbyParams.LobbyId})
		return
	}

	//only players who are registered in the lobby should be able to start the game.
	validPlayer := false
	for _, player := range lobby.Players {
		if player.SessionID == lobbyParams.SessionId {
			validPlayer = true
			break
		}
	}
	if !validPlayer {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Player session is not in the lobby: " + lobbyParams.SessionId})
		return
	}

	err := lobby.StartGame(gs.Questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start game: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"countdownMs": lobby.Countdown, "questionCount": lobby.QuestionCount})
}
