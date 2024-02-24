package main

import (
	"sync"
	"testing"
	"time"
)

//TestGameWaiting //start a game that has a countdown and then verify that the game is "starting" after telling it to start. and verify getting an error if trying to submit answer at that point.
//TestAllWrongMovesGameAhead //if all players have submitted an answer but nobody got it correct, then move on to the next question.
//TestNotEnoughQuestionsSelectsAllGiven
//TestTooManyQuestionsSelectsTheCorrectCount
//TestPlayerCannotAnswerSameQuestionTwice

// answerQuestions simulates answering questions based on per-question delays and correctness.
func answerQuestions(t *testing.T, gameLobby *GameLobby, playerID string, delays []int, correctness []bool, nospam bool, wg *sync.WaitGroup) {
	defer wg.Done()

	answeredQuestions := make(map[int]bool) // Track which questions have been answered.

	for {
		gameLobby.mutex.Lock()
		currentIndex := gameLobby.CurrentQuestionIndex

		// Break if the game has ended.
		if gameLobby.State == Ended {
			gameLobby.mutex.Unlock()
			break
		}

		//i would like this behavior toggleable so we can try to cheat to verify the server prevents it.
		if nospam {
			//Skip if this question has already been answered.
			if answeredQuestions[currentIndex] {
				gameLobby.mutex.Unlock()
				// Break if all questions have been answered
				if len(answeredQuestions) >= len(gameLobby.Questions) {
					t.Logf("Player %s, believes they answered all the questions and shouldnt answer them again.", playerID)
					break
				}
				t.Logf("Player %s, believes they already answered this particular question index %d, and should see if a new one is available after a little nap", playerID, currentIndex)
				time.Sleep(1 * time.Millisecond) // Wait before checking again.
				continue
			}
		}

		currentQuestionID := gameLobby.Questions[currentIndex].ID
		correctIndex := gameLobby.Questions[currentIndex].CorrectIndex
		gameLobby.mutex.Unlock()

		// Wait the specified delay before answering the current question.
		if currentIndex < len(delays) {
			t.Logf("Player %s, believes currentIndex to be %d, go to sleep for %d ms, zzz ....", playerID, currentIndex, delays[currentIndex])
			time.Sleep(time.Duration(delays[currentIndex]) * time.Millisecond)
		} else {
			t.Errorf("Delays array does not cover the question index %d", currentIndex)
			continue
		}

		// Determine answer index based on correctness.
		answerIndex := correctIndex
		if !correctness[currentIndex] {
			// Choose an incorrect answer. Assuming at least 2 options per question. (use modulo to cycle within the bounds, offset by one.)
			answerIndex = (correctIndex + 1) % len(gameLobby.Questions[currentIndex].Options)
		}

		// Submit the answer after the delay.
		t.Logf("Player %s, believes currentIndex to be %d - submitting answer index %d (correct? %t) for this question with ID %s", playerID, currentIndex, answerIndex, correctness[currentIndex], currentQuestionID)
		err, _ := gameLobby.SubmitAnswer(playerID, currentQuestionID, answerIndex)
		answeredQuestions[currentIndex] = true // Mark this question as answered.
		if err != nil {
			t.Logf("Player %s failed to submit answer for question %s: %v", playerID, currentQuestionID, err)
			continue
		}

	}
}

func setupAndStartGame(t *testing.T, questionCount int, countdown int, questions []*Question) *GameLobby {
	lobby := NewGameLobby(questionCount, countdown)

	// Add players
	lobby.AddPlayer("player1")
	lobby.AddPlayer("player2")

	// Start the game with the provided questions
	lobby.StartGame(questions)

	// Adjust the sleep time based on the countdown plus an additional millisecond
	time.Sleep(time.Duration(countdown+1) * time.Millisecond)

	// Check if the game has started as expected
	if lobby.State != Started {
		t.Errorf("Game had %dms countdown so it should be started after waiting %dms.", countdown, countdown+1)
	}

	// Return the lobby instance for further use
	return lobby
}

func TestCompleteGamePlayer2Wins(t *testing.T) {
	// Initialize game with 3 questions, 0ms countdown
	questions := []*Question{
		{ID: "q1", QuestionText: "Question 1", Options: []string{"A", "B", "C"}, CorrectIndex: 1},
		{ID: "q2", QuestionText: "Question 2", Options: []string{"A", "B", "C"}, CorrectIndex: 2},
		{ID: "q3", QuestionText: "Question 3", Options: []string{"A", "B", "C"}, CorrectIndex: 0},
	}
	lobby := setupAndStartGame(t, 3, 0, questions)

	var wg sync.WaitGroup
	wg.Add(2)

	// Player 1 with a delay to simulate slower answering
	go answerQuestions(t, lobby, "player1", []int{100, 100, 100}, []bool{true, true, true}, false, &wg)
	// Player 2 answers much faster
	go answerQuestions(t, lobby, "player2", []int{10, 10, 10}, []bool{true, true, true}, false, &wg)

	wg.Wait()

	// Verify final game state and scores
	if lobby.State != Ended {
		t.Errorf("Expected game state to be Ended, got %v", lobby.State)
	}

	// Assuming we can access players' scores directly; adjust based on your actual data structure
	player1Score := lobby.Players[0].Score
	player2Score := lobby.Players[1].Score
	if player1Score >= player2Score {
		t.Errorf("Expected player 2 to have a higher score. Player 1: %d, Player 2: %d", player1Score, player2Score)
	}
}

func TestPlayer2IsWrong(t *testing.T) {
	// Initialize game with 3 questions, 0ms countdown
	questions := []*Question{
		{ID: "q1", QuestionText: "Question 1", Options: []string{"A", "B", "C"}, CorrectIndex: 1},
		{ID: "q2", QuestionText: "Question 2", Options: []string{"A", "B", "C"}, CorrectIndex: 2},
		{ID: "q3", QuestionText: "Question 3", Options: []string{"A", "B", "C"}, CorrectIndex: 0},
	}
	lobby := setupAndStartGame(t, 3, 0, questions)

	var wg sync.WaitGroup
	wg.Add(2)

	// Player 1 with a delay to simulate slower answering
	go answerQuestions(t, lobby, "player1", []int{100, 100, 100}, []bool{true, true, true}, true, &wg)
	// Player 2 answers much faster (and wrong)
	go answerQuestions(t, lobby, "player2", []int{10, 10, 10}, []bool{false, false, false}, true, &wg)

	wg.Wait()

	// Verify final game state and scores
	if lobby.State != Ended {
		t.Errorf("Expected game state to be Ended, got %v", lobby.State)
	}

	player1Score := lobby.Players[0].Score
	player2Score := lobby.Players[1].Score
	if player2Score >= player1Score {
		t.Errorf("Expected player 2 to have a higher score. Player 1: %d, Player 2: %d", player1Score, player2Score)
	}
}

func TestAddingLobbiesWithAndWithoutPlayer(t *testing.T) {
	// Initialize the Lobbies instance
	lobbies := Lobbies{
		lobbies: make(map[string]*GameLobby),
	}

	// Add the first game lobby without a player
	lobbies.AddLobby(3, 100, nil)

	// Verify there is 1 game in the lobbies with no players
	if len(lobbies.lobbies) != 1 {
		t.Fatalf("Expected 1 game in the lobbies, found %d", len(lobbies.lobbies))
	}

	for _, lobby := range lobbies.lobbies {
		if len(lobby.Players) != 0 {
			t.Errorf("Expected 0 players in the first game lobby, found %d", len(lobby.Players))
		}
	}

	// Add a second game lobby with a player
	player := &Player{SessionID: "player1"}
	lobbies.AddLobby(5, 200, player)

	// Verify there are 2 games in the lobbies
	if len(lobbies.lobbies) != 2 {
		t.Fatalf("Expected 2 games in the lobbies, found %d", len(lobbies.lobbies))
	}

	// Verify only one of the lobbies has a player
	playerCount := 0
	for _, lobby := range lobbies.lobbies {
		playerCount += len(lobby.Players)
	}

	if playerCount != 1 {
		t.Errorf("Expected 1 total player across all lobbies, found %d", playerCount)
	}
}
