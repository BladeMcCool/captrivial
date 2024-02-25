package main

import (
	"encoding/json"
	"github.com/ProlificLabs/captrivia/game"
	"github.com/ProlificLabs/captrivia/server"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

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
func setupServer() (*gin.Engine, *server.GameServer, error) {
	questions, err := loadQuestions()
	if err != nil {
		return nil, nil, err
	}

	//sessions := &SessionStore{Sessions: make(map[string]*PlayerSession)}
	lobbies := game.NewLobbies()
	server := server.NewGameServer(questions, lobbies)

	// Create Gin router and setup routes
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	config := cors.DefaultConfig()
	// allow all origins
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	router.POST("/game/newlobby", server.NewLobbyHandler)
	router.GET("/game/joinlobby/:lobbyId", server.JoinLobbyHandler)
	router.GET("/game/status/:lobbyId", server.GameStatusHandler)
	router.POST("/game/start", server.StartGameHandler)
	router.POST("/game/answer", server.AnswerHandler)
	router.GET("/game/events/:lobbyId/:sessionId", server.WsHandler)

	return router, server, nil
}

func loadQuestions() ([]*game.Question, error) {
	// Use the QUESTIONS_FILE environment variable if it exists; otherwise, default to "questions.json"
	filePath := os.Getenv("QUESTIONS_FILE")
	if filePath == "" {
		filePath = "questions.json"
	}
	log.Printf("loading server questions from file: %s", filePath)

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var questions []*game.Question
	if err := json.Unmarshal(fileBytes, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}
