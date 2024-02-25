import React, {useContext} from 'react';
import { AppContext } from '../App';

const PickAnswer = ({ questions, score, lobbySession, playerSession, setScore, setError, setLoading }) => {
    const { API_BASE } = useContext(AppContext);

    const submitAnswer = async (index) => {
        // We are submitting the index
        setLoading(true);
        const currentQuestion = questions[questions.length-1];
        console.log("believes current question id is: ", currentQuestion.id)
        try {
            const res = await fetch(`${API_BASE}/game/answer`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    lobbyId: lobbySession,
                    sessionId: playerSession,
                    questionId: currentQuestion.id, // field name is "id", not "questionId"
                    answer: index,
                }),
            });
            const data = await res.json();
            if (data.points) {
                //TODO points will tell us if we got the question right or not ... do something fancy if so
                setScore(data.score); // Update score from server's response
            }
        } catch (err) {
            setError("Failed to submit answer.");
        }
        setLoading(false);
    };

    return (
        <div>
            {/* Game session UI */}
            <h3>{questions[questions.length - 1]?.questionText}</h3>
            {questions[questions.length - 1]?.options.map((option, index) => (
                <button key={index} onClick={() => submitAnswer(index)} className="option-button">
                    {option}
                </button>
            ))}
            <p className="score">Score: {score}</p>
        </div>
    );
};

export default PickAnswer;