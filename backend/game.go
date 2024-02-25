package main

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"sync"
	"time"
)

// TODO consider about cleanup of lobbies, perhaps when a game ends something can clean it out.
type Lobbies struct {
	mutex   sync.Mutex
	lobbies map[string]*GameLobby
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

// Message struct to encapsulate game messages
type Message interface{}

type Player struct {
	SessionID         string
	Score             int
	QuestionsAnswered []string     //to hold the ids of the questions that the player answered, in case 'no player answers it correctly first', so we have some way to track it.
	MessageChannel    chan Message // Channel for sending messages to the player
}

func (p *Player) SendMessage(message Message) {
	p.MessageChannel <- message
}

func (p *Player) HasAnsweredQuestion(questionID string) bool {
	for _, qId := range p.QuestionsAnswered {
		if qId == questionID {
			return true
		}
	}
	return false
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
		//LobbyId:              generateSessionID(), //probably replace this with something better for both session and lobbyid generation. maybe a uuid
		QuestionCount:        questionCount,
		Countdown:            countdown,
		State:                Waiting,
		Players:              make([]*Player, 0),
		CurrentQuestionIndex: 0,
	}
}

func (g *GameLobby) GetPlayer(sessionID string) (*Player, error) {
	for _, p := range g.Players {
		if p.SessionID == sessionID {
			return p, nil
		}
	}
	return nil, errors.New("player not found")
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

	// Check if the sessionID is already in the list of players
	for _, player := range g.Players {
		if player.SessionID == sessionID {
			return errors.New("player with this sessionID is already added")
		}
	}

	g.Players = append(g.Players, &Player{
		SessionID:         sessionID,
		Score:             0,
		QuestionsAnswered: []string{},
		MessageChannel:    make(chan Message, g.QuestionCount+2), // Plus 2 for start and finish messages
	})
	// TODO dont forget to close(player.MessageChannel) at some point (how about when the game is done?)
	return nil
}

func (g *GameLobby) StartGame(questionPool []*Question) error {
	g.mutex.Lock()
	if g.State != Waiting {
		return errors.New("Game already started")
	}
	g.State = Starting
	g.mutex.Unlock()

	// Set the shuffled questions for the game
	g.setShuffledQuestionsFromPool(questionPool)

	go func() {
		// Notification mechanism to connected clients (elaborated below)
		for _, player := range g.Players {
			player.SendMessage(map[string]interface{}{
				"countdownMs": 5000,
			})
		}
		time.Sleep(time.Duration(g.Countdown) * time.Millisecond)
		g.mutex.Lock()
		g.State = Started
		g.mutex.Unlock()

		// and how exactly is the question getting in front of the player now?
		g.SendCurrentQuestion()
	}()
	return nil
}

func (g *GameLobby) SendCurrentQuestion() {
	question := g.Questions[g.CurrentQuestionIndex]
	log.Printf("sending next question to all players ...., question id is: %s", question.ID)
	questionForPlayer := map[string]interface{}{ //suppress the correct answer.
		"id":           question.ID,
		"options":      question.Options,
		"questionText": question.QuestionText,
	}
	for _, player := range g.Players {
		player.SendMessage(map[string]interface{}{
			"question": questionForPlayer,
		})
	}
}

func (g *GameLobby) SendGameOver() {
	for _, player := range g.Players {
		player.SendMessage(map[string]bool{
			"gameOver": true,
		})
	}

}

func (g *GameLobby) SubmitAnswer(playerSessionID string, questionID string, answerIndex int) (error, int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.State == Ended {
		return errors.New("game has already ended"), 0
	}
	if g.State != Started {
		return errors.New("game is not started"), 0
	}

	currentQuestion := g.Questions[g.CurrentQuestionIndex]
	if questionID != currentQuestion.ID {
		return errors.New("incorrect question ID"), 0 //should only be trying to answer the question that is currently in front of all players.
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
		log.Printf("pnf!")
		return errors.New("player not found"), 0
	}
	log.Printf("here1")
	// If this player already recorded an answer for this question, then reject this answer.
	if player.HasAnsweredQuestion(questionID) {
		return errors.New("player already answered this question"), 0
	}

	// Record the fact that this player answered this question.
	player.QuestionsAnswered = append(player.QuestionsAnswered, questionID)

	// Validate the answer
	if answerIndex != currentQuestion.CorrectIndex {
		if !g.allPlayersAnswered(questionID) {
			return errors.New("incorrect answer"), 0
		} else {
			return errors.New("incorrect answer (from all players now)"), 0
		}
	}

	// Answer is correct, update player's score
	// using the same number of points that was coded in the original http handler for correct answer.
	awardedPoints := 10
	player.Score += awardedPoints

	// Check if the game has ended and update its state if so.
	g.setNextQuestionOrEndGame()
	return nil, awardedPoints
}

func (g *GameLobby) allPlayersAnswered(questionID string) bool {
	allPlayersAnswered := true
	for _, p := range g.Players {
		if !p.HasAnsweredQuestion(questionID) {
			allPlayersAnswered = false
			break
		}
	}
	if allPlayersAnswered {
		g.setNextQuestionOrEndGame()
	}
	return allPlayersAnswered
}

func (g *GameLobby) setNextQuestionOrEndGame() {
	// Increment the current question index or end the game if all questions are answered
	if g.CurrentQuestionIndex < len(g.Questions)-1 {
		g.CurrentQuestionIndex++
		g.SendCurrentQuestion()
	} else {
		// This was the last question, so end the game
		g.CurrentQuestionIndex = 0
		g.State = Ended
		g.SendGameOver()
	}
}

type GameStatusResult struct {
	State        GameState `json:"state"`
	WinningScore int       `json:"winningScore"`
	Winners      []string  `json:"winners"` // Session IDs of the winning player(s)
}

func (g *GameLobby) GameStatus() GameStatusResult {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	var result GameStatusResult
	result.State = g.State
	result.WinningScore = 0
	scoreToSessions := make(map[int][]string) // Map scores to session IDs

	for _, player := range g.Players {
		score := player.Score
		scoreToSessions[score] = append(scoreToSessions[score], player.SessionID)

		// Update the high score if this player's score is higher
		if score > result.WinningScore {
			result.WinningScore = score
		}
	}

	// Collect session IDs of all players with the high score
	result.Winners = scoreToSessions[result.WinningScore]

	return result
}
