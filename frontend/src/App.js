import React, { useEffect, useState } from "react";
import "./App.css";

// Use REACT_APP_BACKEND_URL or http://localhost:8080 as the API_BASE
const API_BASE = process.env.REACT_APP_BACKEND_URL || "http://localhost:8080";

function App() {
  const [gameSession, setGameSession] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [score, setScore] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [serverMessage, setServerMessage] = useState('');

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = () => {
      console.log('WebSocket Connected');
    };

    ws.onmessage = (event) => {
      setServerMessage(event.data);
    };

    return () => {
      console.log('calling ws.close() ...');
      ws.close();
    };
  }, []);

  const startGame = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/game/start`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });
      const data = await res.json();
      setGameSession(data.sessionId);
      fetchQuestions();
    } catch (err) {
      setError("Failed to start game.");
    }
    setLoading(false);
  };

  const fetchQuestions = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/questions`);
      const data = await res.json();
      setQuestions(data);
    } catch (err) {
      setError("Failed to fetch questions.");
    }
    setLoading(false);
  };

  const submitAnswer = async (index) => {
    // We are submitting the index
    setLoading(true);
    const currentQuestion = questions[currentQuestionIndex];
    try {
      const res = await fetch(`${API_BASE}/answer`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          sessionId: gameSession,
          questionId: currentQuestion.id, // field name is "id", not "questionId"
          answer: index,
        }),
      });
      const data = await res.json();
      if (data.correct) {
        setScore(data.currentScore); // Update score from server's response
      }
      if (currentQuestionIndex < questions.length - 1) {
        setCurrentQuestionIndex(currentQuestionIndex + 1);
      } else {
        endGame();
      }
    } catch (err) {
      setError("Failed to submit answer.");
    }
    setLoading(false);
  };

  const endGame = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/game/end`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          sessionId: gameSession, // need to provide the sessionId
        }),
      });
      const data = await res.json();
      alert(`Game over! Your score: ${data.finalScore}`); // Use the finalScore from the response
      setGameSession(null);
      setQuestions([]);
      setCurrentQuestionIndex(0);
      setScore(0);
    } catch (err) {
      setError("Failed to end game.");
    }
    setLoading(false);
  };

  if (error) return <div className="error">Error: {error}</div>;
  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="App">
      {/* New WebSocket message display */}
      <div className="websocket-panel">
        Server Says: {serverMessage}
      </div>

      {!gameSession ? (
        <button onClick={startGame}>Start Game</button>
      ) : (
        <div>
          <h3>{questions[currentQuestionIndex]?.questionText}</h3>
          {questions[currentQuestionIndex]?.options.map((option, index) => (
            <button
              key={index} // Key should be unique for each child in a list, use index as the key
              onClick={() => submitAnswer(index)} // Pass index instead of option
              className="option-button"
            >
              {option}
            </button>
          ))}
          <p>yay frontend</p>
          <p className="score">Score: {score}</p>
        </div>
      )}
    </div>
  );
}

export default App;
