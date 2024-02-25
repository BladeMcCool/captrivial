package game

import (
	"github.com/google/uuid"
	"sync"
)

// TODO consider about cleanup of lobbies, perhaps when a game ends something can clean it out.
type Lobbies struct {
	mutex   sync.Mutex
	lobbies map[string]*GameLobby
}

// NewLobbies creates and returns a new Lobbies instance
func NewLobbies() *Lobbies {
	return &Lobbies{
		lobbies: make(map[string]*GameLobby),
	}
}

// GetLobby attempts to find and return a lobby by its ID.
// Returns a pointer to the GameLobby and a boolean indicating whether the lobby was found.
func (l *Lobbies) GetLobby(lobbyId string) (*GameLobby, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	lobby, found := l.lobbies[lobbyId]
	return lobby, found
}

func (l *Lobbies) AddLobby(questionCount, countdown int, player *Player) string {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Generate a unique ID for the new lobby
	newLobbyID := uuid.New().String()

	// Create a new GameLobby instance
	newLobby := NewGameLobby(questionCount, countdown)

	// If a player instance is provided, add the player to the new lobby
	if player != nil {
		newLobby.AddPlayer(player.SessionID)
	}

	// Add the new lobby to the lobbies map
	l.lobbies[newLobbyID] = newLobby
	return newLobbyID
}
