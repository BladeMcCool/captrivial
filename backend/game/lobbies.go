package game

import (
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

// TODO consider about cleanup of lobbies, perhaps when a game ends something can clean it out.
type Lobbies struct {
	mutex           sync.Mutex
	lobbies         map[string]*GameLobby
	cleanupInterval time.Duration // Cleanup interval in minutes
}

// NewLobbies creates and returns a new Lobbies instance
func NewLobbies(cleanupInterval time.Duration) *Lobbies {
	return &Lobbies{
		lobbies:         make(map[string]*GameLobby),
		cleanupInterval: cleanupInterval,
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

func (l *Lobbies) StartCleanupRoutine() {
	log.Printf("starting cleanup routine ...")
	go func() {
		for {
			time.Sleep(1 * time.Minute) // Run this check every minute
			l.cleanupExpiredLobbies()
		}
	}()
}

func (l *Lobbies) cleanupExpiredLobbies() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	log.Printf("running cleanup routine ...")
	for id, lobby := range l.lobbies {
		lobby.mutex.Lock()
		log.Printf("running cleanup routine ... checking on lobby id %s", id)
		// Determine if the lobby is expired
		if time.Since(lobby.LastGameInteraction) > l.cleanupInterval {
			log.Printf("removing old lobby uuid %s, closing message channels of %d gamelobby players", id, len(lobby.Players))
			// Perform cleanup for this lobby
			// This should include closing player channels, removing players, etc.
			// For example, closing player channels (simplified):
			for _, player := range lobby.Players {
				close(player.MessageChannel)
			}
			// Remove the lobby from the map
			delete(l.lobbies, id)
		}
		lobby.mutex.Unlock()
	}
}
