package server

import (
	"crypto/rand"
	"fmt"
	"github.com/ProlificLabs/captrivia/game"
)

type GameServer struct {
	Questions []*game.Question
	//Sessions  *SessionStore
	Lobbies *game.Lobbies
}

func NewGameServer(questions []*game.Question, lobbies *game.Lobbies) *GameServer {
	return &GameServer{
		Questions: questions,
		//Sessions:  store,
		Lobbies: lobbies,
	}
}

func (gs *GameServer) generateSessionID() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
