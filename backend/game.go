package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Lobbies struct {
	mutex   sync.Mutex
	lobbies map[string]*GameLobby
}

type Player struct {
	SessionID string
	Score     int
}

type GameState int

const (
	Waiting GameState = iota
	Starting
	Started
	Ended
)

type GameLobby struct {
	mutex                sync.Mutex
	LobbyId              string
	QuestionCount        int
	Countdown            int // milliseconds
	State                GameState
	Players              []*Player
	CurrentQuestionIndex int
	Questions            []*Question
}

func NewGameLobby(questionCount, countdown int) *GameLobby {
	return &GameLobby{
		LobbyId:              generateSessionID(), //probably replace this with something better for both session and lobbyid generation. maybe a uuid
		QuestionCount:        questionCount,
		Countdown:            countdown,
		State:                Waiting,
		Players:              make([]*Player, 0),
		CurrentQuestionIndex: 0,
	}
}

func (g *GameLobby) setShuffledQuestionsFromPool(questions []*Question) {
	//rand.Seed(time.Now().UnixNano()) //apparently the need to do this is gone as of golang 1.20
	shuffled := make([]*Question, len(questions))
	copy(shuffled, questions)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// If count is 0 or exceeds the length of the question pool, use the entire shuffled pool
	if g.QuestionCount == 0 || g.QuestionCount > len(shuffled) {
		g.Questions = shuffled
	}

	// Use the first 'g.QuestionCount' elements of the shuffled slice
	g.Questions = shuffled[:g.QuestionCount]
}

func (g *GameLobby) AddPlayer(sessionID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.State != Waiting {
		return errors.New("cannot add player, lobby is not in waiting state")
	}

	g.Players = append(g.Players, &Player{SessionID: sessionID, Score: 0})
	return nil
}

func (g *GameLobby) StartGame(questionPool []*Question) {
	g.mutex.Lock()
	g.State = Starting
	g.mutex.Unlock()

	// Set the shuffled questions for the game
	g.setShuffledQuestionsFromPool(questionPool)

	go func() {
		time.Sleep(time.Duration(g.Countdown) * time.Millisecond)
		g.mutex.Lock()
		g.State = Started
		g.mutex.Unlock()

		// Notification mechanism to connected clients (elaborated below)
	}()
}

func (g *GameLobby) SubmitAnswer(playerSessionID string, questionID string, answerIndex int) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.State == Ended {
		return errors.New("game has already ended")
	}
	if g.State != Started {
		return errors.New("game is not started")
	}

	currentQuestion := g.Questions[g.CurrentQuestionIndex]
	if questionID != currentQuestion.ID {
		return errors.New("incorrect question ID") //should only be trying to answer the question that is currently in front of all players.
	}

	// Find the player
	var player *Player
	for _, p := range g.Players {
		if p.SessionID == playerSessionID {
			player = p
			break
		}
	}

	if player == nil {
		return errors.New("player not found")
	}

	// Validate the answer
	if answerIndex != currentQuestion.CorrectIndex {
		return errors.New("incorrect answer")
	}

	// Answer is correct, update player's score
	// the requirements suggests adding "points" :D - perhaps in future each question could have its own number of points so that harder questions are worth more than easier ones?
	player.Score += 2

	// Check if the game has ended and update its state if so.
	g.checkIfGameEnded()
	return nil
}

func (g *GameLobby) checkIfGameEnded() {
	// Increment the current question index or end the game if all questions are answered
	if g.CurrentQuestionIndex < len(g.Questions)-1 {
		g.CurrentQuestionIndex++
	} else {
		// This was the last question, so end the game
		g.CurrentQuestionIndex = 0
		g.State = Ended
	}
}
