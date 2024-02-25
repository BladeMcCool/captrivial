package game

type Player struct {
	SessionID         string
	Score             int
	QuestionsAnswered []string     //to hold the ids of the questions that the player answered, in case 'no player answers it correctly first', so we have some way to track it.
	MessageChannel    chan Message // Channel for sending messages to the player
}

// Message struct to encapsulate game messages
type Message interface{}

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
