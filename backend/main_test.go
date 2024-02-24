package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var testRouter *gin.Engine
var testHttpServer *httptest.Server
var testGameServer *GameServer

// TestMain is called before any test runs.
// It allows us to set up things and also clean up after all tests have been run.
func TestMain(m *testing.M) {
	// Set Gin to test mode so that it doesn't print out debug info and we can use testing shortcuts
	gin.SetMode(gin.TestMode)

	var err error
	testRouter, testGameServer, err = setupServer() // This should call the same setupServer which is used in main.
	if err != nil {
		log.Fatal("Failed to set up test server:", err)
	}

	// Start a new httptest server using the testRouter.
	testHttpServer = httptest.NewServer(testRouter)

	runTests := m.Run()

	// Close the test server
	testHttpServer.Close()

	// Exit with the result of the test suite run
	os.Exit(runTests)
}

// TODO use proper json serialization to submit the params to the handlers.
func TestNewLobbyHandler(t *testing.T) {
	resp, err := http.Post(testHttpServer.URL+"/game/newlobby", "application/json", strings.NewReader(fmt.Sprintf(`{"questionCount":%d, "countdownMs":%d}`, 5, 100)))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	// Decode JSON response
	var response struct {
		LobbyId string `json:"lobbyId"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Check if a lobbyid as been returned
	if response.LobbyId == "" {
		t.Errorf("Response should contain a lobbyId")
	}

	// Validate the lobbyId is a UUID
	_, err = uuid.Parse(response.LobbyId)
	if err != nil {
		t.Errorf("The lobbyId '%s' is not a valid UUID: %v", response.LobbyId, err)
	}
}

func TestJoinLobbyHandler(t *testing.T) {
	resp, err := http.Post(testHttpServer.URL+"/game/newlobby", "application/json", strings.NewReader(fmt.Sprintf(`{"questionCount":%d, "countdownMs":%d}`, 5, 100)))
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	// Decode JSON response
	var response struct {
		LobbyId string `json:"lobbyId"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Check if a sessionId as been returned
	if response.LobbyId == "" {
		t.Errorf("Response should contain a lobbyId")
	}

	// Now join this lobby as a 2nd player.
	resp, err = http.Get(testHttpServer.URL + "/game/joinlobby/" + response.LobbyId)
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}
}

// test a full game
func TestFullGameSinglePlayer(t *testing.T) {
	// Start a new game
	resp, err := http.Post(testHttpServer.URL+"/game/newlobby", "application/json", strings.NewReader(fmt.Sprintf(`{"questionCount":%d, "countdownMs":%d}`, 3, 100)))
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	// Decode JSON response
	var response struct {
		LobbyId   string `json:"lobbyId"`
		SessionId string `json:"sessionId"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Check if a lobbyid as been returned
	if response.LobbyId == "" {
		t.Errorf("Response should contain a lobbyId")
	}
	if response.SessionId == "" {
		t.Errorf("Response should contain a sessionId")
	}

	resp, err = http.Post(testHttpServer.URL+"/game/start", "application/json", strings.NewReader(fmt.Sprintf(`{"lobbyId":"%s"}`, response.LobbyId)))
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to start a new game: %v", err)
	}

	// Check for the correct status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	// JSON response will tell us when the game is starting (mostly just informational since we should be subscribing to an event stream of some kind to actually get the first question)

	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	t.Fatalf("Failed to read response body: %v", err)
	//}
	//bodyText := string(bodyBytes)
	//fmt.Println("Response body:", bodyText)

	var startResponse struct {
		CountdownMs *int `json:"countdownMs"`
	}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
	t.Logf("Game starts in %d ms", *startResponse.CountdownMs)

	// Check if a CountdownMs has been returned
	if startResponse.CountdownMs == nil {
		t.Errorf("Response should contain a CountdownMs")
	}

	//attempt to just submit some answer to the first question, even though the game isn't started. should get some kind of error response.
	resp, err = http.Post(testHttpServer.URL+"/game/answer", "application/json", strings.NewReader(fmt.Sprintf(`{"sessionId":"%s", "lobbyId":"%s", "questionId":"%s", "answer":%d}`, response.SessionId, response.LobbyId, "", 0)))
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyText := string(bodyBytes)
	fmt.Println("Response body:", bodyText)

	//sleep for the duration until we know the game should be started.
	time.Sleep(time.Duration(*startResponse.CountdownMs) * time.Millisecond)

	//not implementing the streaming event recovery here just at the moment, lets inspect the lobby/game to determine question/answer to use for a happy path.
	//we will just submit the 3 correct answers in a row back to back and confirm that the game is now ended and that we got the 6 points.
	lobby, found := testGameServer.Lobbies.GetLobby(response.LobbyId)
	if !found {
		t.Fatalf("expected lobby %s was not found", response.LobbyId)
	}

	submitAnswer := lobby.Questions[lobby.CurrentQuestionIndex].CorrectIndex
	submitQuestionId := lobby.Questions[lobby.CurrentQuestionIndex].ID
	resp, err = http.Post(testHttpServer.URL+"/game/answer", "application/json", strings.NewReader(fmt.Sprintf(`{"sessionId":"%s", "lobbyId":"%s", "questionId":"%s", "answer":%d}`, response.SessionId, response.LobbyId, submitQuestionId, submitAnswer)))
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyText = string(bodyBytes)
	fmt.Println("Response body:", bodyText)

	//
	//var startGameResponse map[string]string
	//err = json.NewDecoder(resp.Body).Decode(&startGameResponse)
	//if err != nil {
	//	t.Fatalf("Failed to decode JSON response: %v", err)
	//}
	//sessionID, exists := startGameResponse["sessionId"]
	//if !exists {
	//	t.Fatalf("Response does not contain 'sessionId'")
	//}
	//
	//// Get questions
	//resp, err = http.Get(testHttpServer.URL + "/questions")
	//if err != nil {
	//	t.Fatalf("Failed to get questions: %v", err)
	//}
	//defer resp.Body.Close()
	//
	//// Check for the correct status code
	//if resp.StatusCode != http.StatusOK {
	//	t.Errorf("Expected status OK; got %v", resp.Status)
	//}
	//
	//// Decode JSON response to get the questions
	//var questions []Question
	//err = json.NewDecoder(resp.Body).Decode(&questions)
	//if err != nil {
	//	t.Fatalf("Failed to decode JSON response: %v", err)
	//}
	//if len(questions) == 0 {
	//	t.Fatalf("No questions received")
	//}
	//
	//// Answer each question (assuming the answer is always the first option)
	//for _, question := range questions {
	//	// Make sure we haven't been given the answer.  We're using the same struct here for the server-side
	//	// handler and the "client", so if it wasn't set it should always be 0
	//	if question.CorrectIndex != 0 {
	//		t.Fatalf("Backend returned answer index")
	//	}
	//
	//	answerPayload := fmt.Sprintf(`{"sessionId":"%s", "questionId":"%s", "answer":%d}`, sessionID, question.ID, 0)
	//	answerReader := strings.NewReader(answerPayload)
	//	resp, err = http.Post(testHttpServer.URL+"/answer", "application/json", answerReader)
	//	if err != nil {
	//		t.Fatalf("Failed to post answer: %v", err)
	//	}
	//	defer resp.Body.Close()
	//
	//	// Check for the correct status code
	//	if resp.StatusCode != http.StatusOK {
	//		t.Errorf("Expected status OK; got %v", resp.Status)
	//	}
	//
	//	// Decode JSON response to check if the answer was correct
	//	var answerResponse map[string]interface{}
	//	err = json.NewDecoder(resp.Body).Decode(&answerResponse)
	//	if err != nil {
	//		t.Fatalf("Failed to decode JSON response: %v", err)
	//	}
	//	if _, exists := answerResponse["correct"]; !exists {
	//		t.Errorf("Response should contain 'correct' field")
	//	}
	//}
	//
	//// End the game
	//endGamePayload := fmt.Sprintf(`{"sessionId":"%s"}`, sessionID)
	//endGameReader := strings.NewReader(endGamePayload)
	//resp, err = http.Post(testHttpServer.URL+"/game/end", "application/json", endGameReader)
	//if err != nil {
	//	t.Fatalf("Failed to end the game: %v", err)
	//}
	//defer resp.Body.Close()
	//
	//// Check for the correct status code
	//if resp.StatusCode != http.StatusOK {
	//	t.Errorf("Expected status OK; got %v", resp.Status)
	//}
	//
	//// Decode JSON response to check the final score
	//var endGameResponse map[string]interface{}
	//err = json.NewDecoder(resp.Body).Decode(&endGameResponse)
	//if err != nil {
	//	t.Fatalf("Failed to decode JSON response: %v", err)
	//}
	//if _, exists := endGameResponse["finalScore"]; !exists {
	//	t.Errorf("Response should contain 'finalScore' field")
	//}
}

func TestFullGameMultiPlayer(t *testing.T) {
	//	// Start a new lobby
	//	resp, err := http.Post(testHttpServer.URL+"/game/newlobby", "application/json", nil)
	//	if err != nil {
	//		t.Fatalf("Failed to start a new lobby: %v", err)
	//	}
	//	defer resp.Body.Close()
	//
	//	// Check for the correct status code
	//	if resp.StatusCode != http.StatusOK {
	//		t.Errorf("Expected status OK; got %v", resp.Status)
	//	}
	//
	//	// Decode JSON response to get the lobby ID
	//	var startLobbyResponse map[string]string
	//	err = json.NewDecoder(resp.Body).Decode(&startLobbyResponse)
	//	if err != nil {
	//		t.Fatalf("Failed to decode JSON response: %v", err)
	//	}
	//	sessionID, exists := startGameResponse["sessionId"]
	//	if !exists {
	//		t.Fatalf("Response does not contain 'sessionId'")
	//	}
	//
	//	// Start a new game
	//	resp, err := http.Post(testHttpServer.URL+"/game/start", "application/json", nil)
	//	if err != nil {
	//		t.Fatalf("Failed to start a new game: %v", err)
	//	}
	//	defer resp.Body.Close()
	//
	//	// Check for the correct status code
	//	if resp.StatusCode != http.StatusOK {
	//		t.Errorf("Expected status OK; got %v", resp.Status)
	//	}
	//
	//	// Decode JSON response to get the session ID
	//	var startGameResponse map[string]string
	//	err = json.NewDecoder(resp.Body).Decode(&startGameResponse)
	//	if err != nil {
	//		t.Fatalf("Failed to decode JSON response: %v", err)
	//	}
	//	sessionID, exists := startGameResponse["sessionId"]
	//	if !exists {
	//		t.Fatalf("Response does not contain 'sessionId'")
	//	}
	//
	//	// Get questions
	//	resp, err = http.Get(testHttpServer.URL + "/questions")
	//	if err != nil {
	//		t.Fatalf("Failed to get questions: %v", err)
	//	}
	//	defer resp.Body.Close()
	//
	//	// Check for the correct status code
	//	if resp.StatusCode != http.StatusOK {
	//		t.Errorf("Expected status OK; got %v", resp.Status)
	//	}
	//
	//	// Decode JSON response to get the questions
	//	var questions []Question
	//	err = json.NewDecoder(resp.Body).Decode(&questions)
	//	if err != nil {
	//		t.Fatalf("Failed to decode JSON response: %v", err)
	//	}
	//	if len(questions) == 0 {
	//		t.Fatalf("No questions received")
	//	}
	//
	//	// Answer each question (assuming the answer is always the first option)
	//	for _, question := range questions {
	//		// Make sure we haven't been given the answer.  We're using the same struct here for the server-side
	//		// handler and the "client", so if it wasn't set it should always be 0
	//		if question.CorrectIndex != 0 {
	//			t.Fatalf("Backend returned answer index")
	//		}
	//
	//		answerPayload := fmt.Sprintf(`{"sessionId":"%s", "questionId":"%s", "answer":%d}`, sessionID, question.ID, 0)
	//		answerReader := strings.NewReader(answerPayload)
	//		resp, err = http.Post(testHttpServer.URL+"/answer", "application/json", answerReader)
	//		if err != nil {
	//			t.Fatalf("Failed to post answer: %v", err)
	//		}
	//		defer resp.Body.Close()
	//
	//		// Check for the correct status code
	//		if resp.StatusCode != http.StatusOK {
	//			t.Errorf("Expected status OK; got %v", resp.Status)
	//		}
	//
	//		// Decode JSON response to check if the answer was correct
	//		var answerResponse map[string]interface{}
	//		err = json.NewDecoder(resp.Body).Decode(&answerResponse)
	//		if err != nil {
	//			t.Fatalf("Failed to decode JSON response: %v", err)
	//		}
	//		if _, exists := answerResponse["correct"]; !exists {
	//			t.Errorf("Response should contain 'correct' field")
	//		}
	//	}
	//
	//	// End the game
	//	endGamePayload := fmt.Sprintf(`{"sessionId":"%s"}`, sessionID)
	//	endGameReader := strings.NewReader(endGamePayload)
	//	resp, err = http.Post(testHttpServer.URL+"/game/end", "application/json", endGameReader)
	//	if err != nil {
	//		t.Fatalf("Failed to end the game: %v", err)
	//	}
	//	defer resp.Body.Close()
	//
	//	// Check for the correct status code
	//	if resp.StatusCode != http.StatusOK {
	//		t.Errorf("Expected status OK; got %v", resp.Status)
	//	}
	//
	//	// Decode JSON response to check the final score
	//	var endGameResponse map[string]interface{}
	//	err = json.NewDecoder(resp.Body).Decode(&endGameResponse)
	//	if err != nil {
	//		t.Fatalf("Failed to decode JSON response: %v", err)
	//	}
	//	if _, exists := endGameResponse["finalScore"]; !exists {
	//		t.Errorf("Response should contain 'finalScore' field")
	//	}
}
