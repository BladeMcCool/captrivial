package server

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to find lobby: " + submittedAnswer.LobbyId})
		return
	}

	err, points := lobby.SubmitAnswer(submittedAnswer.SessionID, submittedAnswer.QuestionID, submittedAnswer.Answer)
	//the errors here can all be treated as non-errors, the important part is whether any points was awarded. we could maybe get more info and track a score but the server is going to keep track and push updates to the client so, not worrying about it here.
	if err != nil {
		log.Printf("sumbissionError: %s", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"submissionError": err.Error(),
		})
		return
	}
	player, err := lobby.GetPlayer(submittedAnswer.SessionID)
	if err != nil {
		log.Printf("sumbissionError %s", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"submissionError": err.Error(),
		})
		return
	}

	log.Printf("respond with points %d", points)
	c.JSON(http.StatusOK, gin.H{
		"points": points,
		"score":  player.Score,
	})
}

//func (gs *GameServer) QuestionsHandler(c *gin.Context) {
//	shuffledQuestions := shuffleQuestions(gs.Questions)
//	c.JSON(http.StatusOK, shuffledQuestions[:10])
//}

//func (gs *GameServer) EndGameHandler(c *gin.Context) {
//	var request struct {
//		SessionID string `json:"sessionId"`
//	}
//	if err := c.BindJSON(&request); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	session, exists := gs.Sessions.GetSession(request.SessionID)
//	if !exists {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"finalScore": session.Score})
//}

//
//func (gs *GameServer) checkAnswer(questionID string, submittedAnswer int) (bool, error) {
//	for _, question := range gs.Questions {
//		if question.ID == questionID {
//			return question.CorrectIndex == submittedAnswer, nil
//		}
//	}
//	return false, errors.New("question not found")
//}
//
//func shuffleQuestions(questions []*Question) []Question {
//	rand.Seed(time.Now().UnixNano())
//	qs := make([]Question, len(questions))
//
//	// Copy the questions manually, instead of with copy(), so that we can remove
//	// the CorrectIndex property
//	for i, q := range questions {
//		qs[i] = Question{ID: q.ID, QuestionText: q.QuestionText, Options: q.Options}
//	}
//
//	rand.Shuffle(len(qs), func(i, j int) {
//		qs[i], qs[j] = qs[j], qs[i]
//	})
//	return qs
//}
