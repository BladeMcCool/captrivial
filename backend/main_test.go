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

	resp, err = http.Post(testHttpServer.URL+"/game/start", "application/json", strings.NewReader(fmt.Sprintf(`{"lobbyId":"%s","sessionId":"%s"}`, response.LobbyId, response.SessionId)))
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to start a new game: %v", err)
	}

	// Check for the correct status code
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	// JSON response will tell us when the game is starting (mostly just informational since we should be subscribing to an event stream of some kind to actually get the first question)
	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	t.Fatalf("Failed to read response body: %v", err)
	//}
	//bodyText := string(bodyBytes)
	//fmt.Println("Response body:", bodyText)

	var startResponse struct {
		CountdownMs int `json:"countdownMs"`
	}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
	t.Logf("Game starts in %d ms", startResponse.CountdownMs)

	// a nonzero CountdownMs should be present -- unless we want to support instant start in which case we can remove this sort of check b/c 0 is then valid.
	if startResponse.CountdownMs == 0 {
		t.Errorf("Response should contain a nonzero CountdownMs")
	}

	//attempt to just submit some answer to the first question, even though the game isn't started. should get some kind of error response.
	resp, err = http.Post(testHttpServer.URL+"/game/answer", "application/json", strings.NewReader(fmt.Sprintf(`{"sessionId":"%s", "lobbyId":"%s", "questionId":"%s", "answer":%d}`, response.SessionId, response.LobbyId, "", 0)))
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyText := string(bodyBytes)
	fmt.Println("Response body:", bodyText)
	if !strings.Contains(bodyText, "game is not started") {
		t.Fatalf("should have got an error message about the game not being started yet.")
	}

	//sleep for the duration until we know the game should be started
	time.Sleep(time.Duration(startResponse.CountdownMs) * time.Millisecond)

	//not implementing the streaming event recovery here just at the moment, lets inspect the lobby/game to determine question/answer to use for a happy path.
	//we will just submit the 3 correct answers in a row back to back and confirm that the game is now ended and that we got the 6 points.
	lobby, found := testGameServer.Lobbies.GetLobby(response.LobbyId)
	if !found {
		t.Fatalf("expected lobby %s was not found", response.LobbyId)
	}

	for i := 0; i < 3; i++ {
		var submitAnswerResponse struct {
			Points int `json:"points"`
		}
		submitAnswer := lobby.Questions[lobby.CurrentQuestionIndex].CorrectIndex
		submitQuestionId := lobby.Questions[lobby.CurrentQuestionIndex].ID
		resp, err = http.Post(testHttpServer.URL+"/game/answer", "application/json", strings.NewReader(fmt.Sprintf(`{"sessionId":"%s", "lobbyId":"%s", "questionId":"%s", "answer":%d}`, response.SessionId, response.LobbyId, submitQuestionId, submitAnswer)))
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&submitAnswerResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}
		if submitAnswerResponse.Points != 10 {
			t.Fatalf("failed to get awarded 10 points for our answer.")
		}
	}

	//call to the game status endpoint just to confirm we got the expected result
	resp, err = http.Get(testHttpServer.URL + "/game/status/" + response.LobbyId)
	defer resp.Body.Close()

	var gameStatusResponse GameStatusResult
	err = json.NewDecoder(resp.Body).Decode(&gameStatusResponse)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	//at this point the game status should be concluded.
	if gameStatusResponse.State != Ended {
		t.Fatalf("last question was answered and points awarded - game should be over now.")
	}
	if gameStatusResponse.WinningScore != 30 { //3 questions, 10 points each
		t.Fatalf("unexpected winning score")
	}
	//confirm that there is only one winner, and that it is us.
	if len(gameStatusResponse.Winners) != 1 {
		t.Fatalf("wrong number of winners")
	}

	if gameStatusResponse.Winners[0] != response.SessionId {
		t.Fatalf("winner had an unexpected sessionid")
	}
}

// TODO use this in all the tests for new lobby/join lobby response handling.
type joinGameResponse struct {
	LobbyId   string `json:"lobbyId"`
	SessionId string `json:"sessionId"`
}

func TestFullGameMultiPlayer(t *testing.T) {
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
	var response joinGameResponse
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
	player1SessionId := response.SessionId
	_ = player1SessionId

	// Now join this lobby as a 2nd player.
	resp, err = http.Get(testHttpServer.URL + "/game/joinlobby/" + response.LobbyId)
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}
	var p2response joinGameResponse
	err = json.NewDecoder(resp.Body).Decode(&p2response)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
	t.Logf("%+v", p2response)
	player2SessionId := p2response.SessionId

	resp, err = http.Post(testHttpServer.URL+"/game/start", "application/json", strings.NewReader(fmt.Sprintf(`{"lobbyId":"%s","sessionId":"%s"}`, response.LobbyId, player2SessionId)))
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to start a new game: %v", err)
	}

	// Check for the correct status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	var startResponse struct {
		CountdownMs int `json:"countdownMs"`
	}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
	t.Logf("Game starts in %d ms", startResponse.CountdownMs)

	//sleep for the duration until we know the game should be started.
	extraSleepMs := 50 //so that goroutines have time to wake up and do stuff (i think that is what is goin on anyway)
	time.Sleep(time.Duration(startResponse.CountdownMs+extraSleepMs) * time.Millisecond)
	lobby, found := testGameServer.Lobbies.GetLobby(response.LobbyId)
	if !found {
		t.Fatalf("expected lobby %s was not found", response.LobbyId)
	}
	if lobby.State != Started {
		t.Fatalf("expected lobby to have started the game, but it was in a different state: %d", lobby.State) //TODO implement some stringification for the state enum iota values ...
	}

	//lets have player 2 be the winner by submitting more correct answers than player 1
	submitterSessionIds := []string{player2SessionId, player1SessionId, player2SessionId}
	for i := 0; i < 3; i++ {
		var submitAnswerResponse struct {
			Points int `json:"points"`
		}
		submitAnswer := lobby.Questions[lobby.CurrentQuestionIndex].CorrectIndex
		submitQuestionId := lobby.Questions[lobby.CurrentQuestionIndex].ID
		resp, err = http.Post(testHttpServer.URL+"/game/answer", "application/json", strings.NewReader(fmt.Sprintf(`{"sessionId":"%s", "lobbyId":"%s", "questionId":"%s", "answer":%d}`, submitterSessionIds[i], response.LobbyId, submitQuestionId, submitAnswer)))
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&submitAnswerResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}
		if submitAnswerResponse.Points != 10 {
			t.Fatalf("failed to get awarded 10 points for our answer.")
		}
	}

	//call to the game status endpoint just to confirm we got the expected result
	resp, err = http.Get(testHttpServer.URL + "/game/status/" + response.LobbyId)
	defer resp.Body.Close()

	var gameStatusResponse GameStatusResult
	err = json.NewDecoder(resp.Body).Decode(&gameStatusResponse)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	//at this point the game status should be concluded.
	if gameStatusResponse.State != Ended {
		t.Fatalf("last question was answered and points awarded - game should be over now.")
	}
	if gameStatusResponse.WinningScore != 20 { //player 2 should have got the two that were submitted correct.
		t.Fatalf("unexpected winning score")
	}
	//confirm that there is only one winner, and that it is us.
	if len(gameStatusResponse.Winners) != 1 {
		t.Fatalf("wrong number of winners")
	}
	//and that it is the expected winner.
	if gameStatusResponse.Winners[0] != player2SessionId {
		t.Fatalf("winner had an unexpected sessionid")
	}
}
