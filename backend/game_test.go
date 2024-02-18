package main

import (
	"testing"
	"time"
)

//TestGameWaiting //start a game that has a countdown and then verify that the game is "starting" after telling it to start.
//TestAllWrongMovesGameAhead //if all players have submitted an answer but nobody got it correct, then move on to the next question.

func TestCompleteGame(t *testing.T) {
	// Initialize game with 3 questions, 0ms countdown
	questions := []*Question{
		{ID: "q1", QuestionText: "Question 1", Options: []string{"A", "B", "C"}, CorrectIndex: 1},
		{ID: "q2", QuestionText: "Question 2", Options: []string{"A", "B", "C"}, CorrectIndex: 2},
		{ID: "q3", QuestionText: "Question 3", Options: []string{"A", "B", "C"}, CorrectIndex: 0},
	}
	lobby := NewGameLobby(3, 0)

	// Add players
	lobby.AddPlayer("player1")
	lobby.AddPlayer("player2")
	lobby.StartGame(questions)

	// Ensure game starts
	time.Sleep(1 * time.Millisecond) // Adjust based on actual countdown and processing time
	if lobby.State != Started {
		t.Errorf("Game had 0ms countdown so it should be started after waiting 1ms.")
	}
}
