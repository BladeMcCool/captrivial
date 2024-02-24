package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Question struct {
	ID           string   `json:"id"`
	QuestionText string   `json:"questionText"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correctIndex"`
}

type PlayerSession struct {
	Score int
}

type SessionStore struct {
	sync.Mutex
	Sessions map[string]*PlayerSession
}

func (store *SessionStore) CreateSession() string {
	store.Lock()
	defer store.Unlock()

	uniqueSessionID := generateSessionID()
	store.Sessions[uniqueSessionID] = &PlayerSession{Score: 0}

	return uniqueSessionID
}

func (store *SessionStore) GetSession(sessionID string) (*PlayerSession, bool) {
	store.Lock()
	defer store.Unlock()

	session, exists := store.Sessions[sessionID]
	return session, exists
}

func generateSessionID() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

type GameServer struct {
	Questions []*Question
	Sessions  *SessionStore
	Lobbies   *Lobbies
}

func main() {
	// Setup the server
	router, _, err := setupServer()
	if err != nil {
		log.Fatalf("Server setup failed: %v", err)
	}

	// set port to PORT or 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Println("Server starting on port " + port)
	log.Fatal(router.Run(":" + port))
}

// setupServer configures and returns a new Gin instance with all routes.
// It also returns an error if there is a failure in setting up the server, e.g. loading questions.
func setupServer() (*gin.Engine, *GameServer, error) {
	questions, err := loadQuestions()
	if err != nil {
		return nil, nil, err
	}

	sessions := &SessionStore{Sessions: make(map[string]*PlayerSession)}
	lobbies := &Lobbies{lobbies: make(map[string]*GameLobby)}
	server := NewGameServer(questions, sessions, lobbies)

	// Create Gin router and setup routes
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	config := cors.DefaultConfig()
	// allow all origins
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	// windows cmd curl:
	// curl -v -H "Content-Type: application/json" -d "{\"questionCount\":3,\"countdownMs\":5000}" http://localhost:8080/game/newlobby
	router.POST("/game/newlobby", server.NewLobbyHandler)
	// curl -v http://localhost:8080/game/joinlobby/740ee5bc-0acc-4ff7-9483-4f65ea652638 //using a GET request b/c the url is supposed to be share-able, and i expect ppl to be able to copy/paste it into browser address bar.
	router.GET("/game/joinlobby/:lobbyId", server.JoinLobbyHandler)
	router.GET("/game/status/:lobbyId", server.GameStatusHandler)
	router.POST("/game/start", server.StartGameHandler)
	router.POST("/game/answer", server.AnswerHandler)

	//router.POST("/game/end", server.EndGameHandler)
	//router.GET("/questions", server.QuestionsHandler)
	router.GET("/game/events", server.WsHandler)

	return router, server, nil
}

func NewGameServer(questions []*Question, store *SessionStore, lobbies *Lobbies) *GameServer {
	return &GameServer{
		Questions: questions,
		Sessions:  store,
		Lobbies:   lobbies,
	}
}

func (gs *GameServer) NewLobbyHandler(c *gin.Context) {
	var gameParams struct {
		QuestionCount int `json:"questionCount"`
		CountdownMs   int `json:"countdownMs"`
	}
	if err := c.ShouldBindJSON(&gameParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	sessionID := gs.Sessions.CreateSession()
	lobbyID := gs.Lobbies.AddLobby(gameParams.QuestionCount, gameParams.CountdownMs, &Player{
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
	sessionId := gs.Sessions.CreateSession() //treat as a new player when joining a lobby. session is to identify the player within the lobby.
	err := lobby.AddPlayer(sessionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join lobby: " + err.Error()})
		return
	}

	// Respond with a success message or other relevant information
	c.JSON(http.StatusOK, gin.H{"message": "Joined lobby successfully", "lobbyId": lobbyId, "sessionId": sessionId})
}

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

func (gs *GameServer) QuestionsHandler(c *gin.Context) {
	shuffledQuestions := shuffleQuestions(gs.Questions)
	c.JSON(http.StatusOK, shuffledQuestions[:10])
}

func (gs *GameServer) AnswerHandler(c *gin.Context) {
	var submittedAnswer struct {
		SessionID  string `json:"sessionId"`
		QuestionID string `json:"questionId"`
		LobbyId    string `json:"lobbyId"`
		Answer     int    `json:"answer"`
	}
	if err := c.ShouldBindJSON(&submittedAnswer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	lobby, found := gs.Lobbies.GetLobby(submittedAnswer.LobbyId)
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find lobby: " + submittedAnswer.LobbyId})
		return
	}

	err, points := lobby.SubmitAnswer(submittedAnswer.SessionID, submittedAnswer.QuestionID, submittedAnswer.Answer)
	//the errors here can all be treated as non-errors, the important part is whether any points was awarded. we could maybe get more info and track a score but the server is going to keep track and push updates to the client so, not worrying about it here.
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"submissionError": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"points": points,
	})

}

func (gs *GameServer) EndGameHandler(c *gin.Context) {
	var request struct {
		SessionID string `json:"sessionId"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	session, exists := gs.Sessions.GetSession(request.SessionID)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"finalScore": session.Score})
}

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

func (gs *GameServer) checkAnswer(questionID string, submittedAnswer int) (bool, error) {
	for _, question := range gs.Questions {
		if question.ID == questionID {
			return question.CorrectIndex == submittedAnswer, nil
		}
	}
	return false, errors.New("question not found")
}

func shuffleQuestions(questions []*Question) []Question {
	rand.Seed(time.Now().UnixNano())
	qs := make([]Question, len(questions))

	// Copy the questions manually, instead of with copy(), so that we can remove
	// the CorrectIndex property
	for i, q := range questions {
		qs[i] = Question{ID: q.ID, QuestionText: q.QuestionText, Options: q.Options}
	}

	rand.Shuffle(len(qs), func(i, j int) {
		qs[i], qs[j] = qs[j], qs[i]
	})
	return qs
}

func loadQuestions() ([]*Question, error) {
	fileBytes, err := ioutil.ReadFile("questions.json")
	if err != nil {
		return nil, err
	}

	var questions []*Question
	if err := json.Unmarshal(fileBytes, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

func (gs *GameServer) WsHandler(c *gin.Context) {
	w := c.Writer
	r := c.Request

	// Create a ticker that ticks every second
	ticker := time.NewTicker(100 * time.Millisecond)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}
	somefunc := func() {
		log.Println("close the websocket conn")
		conn.Close()

		log.Println("stop the ticker")
		ticker.Stop()
	}
	defer somefunc()

	spamcount := 0
	for {
		select {
		case <-ticker.C:
			spamcount++
			// Send a message every tick
			message := fmt.Sprintf("Spam message %d from server", spamcount)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Println("Write error:", err)
				//return // Exit the loop and end the connection on error
			}
		}
	}
	//
	//for {
	//	log.Println("send a message down the websocket")
	//	if err := conn.WriteMessage(messageType, p); err != nil {
	//		log.Println("ending the loop due to echoing the message back. wait, wtf message?")
	//		return // Ends the loop if an error occurs
	//	}
	//
	//	//messageType, p, err := conn.ReadMessage()
	//	//if err != nil {
	//	//	log.Printf("ending the loop due to error: %s", err.Error())
	//	//	return // Ends the loop if the connection is closed or an error occurs
	//	//}
	//	//// Echo the received message back to the client
	//	//if err := conn.WriteMessage(messageType, p); err != nil {
	//	//	log.Println("ending the loop due to echoing the message back. wait, wtf message?")
	//	//	return // Ends the loop if an error occurs
	//	//}
	//}
}
