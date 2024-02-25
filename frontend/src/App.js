import React, { useEffect, useState } from "react";
import "./App.css";
import { useParams } from "react-router-dom"; // Import useParams hook

// Use REACT_APP_BACKEND_URL or http://localhost:8080 as the API_BASE
const API_BASE = process.env.REACT_APP_BACKEND_URL || "http://localhost:8080";

function App() {
  const [lobbySession, setLobbySession] = useState(null);
  const [playerSession, setPlayerSession] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [gameParams, setGameParams] = useState({});
  const [gameStarted, setGameStarted] = useState(false);
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [score, setScore] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [ws, setWs] = useState(null); // Store WebSocket connection
  const [serverMessage, setServerMessage] = useState('');

  // Effect for WebSocket setup
  useEffect(() => {
    if (!playerSession || !lobbySession) return; // Only connect WebSocket after lobby is waiting

    const websocket = new WebSocket(`ws://localhost:8080/game/events/${lobbySession}/${playerSession}`);

    websocket.onopen = () => {
      console.log('WebSocket Connected');
    };

    websocket.onmessage = (event) => {
      setServerMessage(event.data);
      const data = JSON.parse(event.data);
      // Handle different types of messages
      if (data.question) {
        setQuestions(prev => [...prev, data.question]);
        setCurrentQuestionIndex(questions.length); // Set to display the new question
      } else if (data.gameOver) {
        alert(`Game over! Winner: ${data.winner}, Score: ${data.score}`);
        // Reset game state here if needed
      }
      // Add more conditions as needed based on your server messages
    };

    setWs(websocket); // Store WebSocket connection

    return () => {
      console.log('Closing WebSocket...');
      websocket.close();
    };
  }, [playerSession, lobbySession, questions.length]); // Re-connect WebSocket if playerSession changes

  // Use useParams hook to extract lobby UUID from the URL
  let { lobbyUuid } = useParams();
  // console.log("hrm")

  // Effect to set lobbySession based on URL lobbyUuid
  useEffect(() => {
    if (lobbyUuid) {
      console.log("there is a lobbyUuid and it is", lobbyUuid)
      setLobbySession(lobbyUuid);
    }
  }, [lobbyUuid]);

  const createNewLobby = async () => {
    setLoading(true);
    setError(null);
    try {
      try {
        const res = await fetch(`${API_BASE}/game/newlobby`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            questionCount: 5,
            countdownMs: 5000,
          }),
        });
        const data = await res.json();
        if (res.ok) {
          console.log("game lobby:", data)
          // Assuming the response includes the lobbySession or playerSession identifier
          // setLobbySession(data.lobbyId); // Update this line based on your actual response structure
          setPlayerSession(data.sessionId);
          setLobbySession(data.lobbyId);
          setGameParams(data);
          // Additional logic to handle successful lobby creation
        } else {
          throw new Error(data.error || "Failed to create new lobby");
        }
      } catch (err) {
        setError(err.message);
      }

      // const res = await fetch(`${API_BASE}/game/newlobby`, {
      //   method: "POST",
      //   headers: {
      //     "Content-Type": "application/json",
      //   },
      // });
      // const data = await res.json();
      // setGameSession(data.sessionId);
      // setLobbySession(data.lobbyId);
      // setGameParams(data);
      // fetchQuestions();
    } catch (err) {
      setError("Failed to create lobby");
    }
    setLoading(false);
  };

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
      setPlayerSession(data.sessionId);
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
          sessionId: playerSession,
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
          sessionId: playerSession, // need to provide the sessionId
        }),
      });
      const data = await res.json();
      alert(`Game over! Your score: ${data.finalScore}`); // Use the finalScore from the response
      setPlayerSession(null);
      setLobbySession(null);
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

  // return (
  //   <div className="App">
  //     {/* New WebSocket message display */}
  //     <div className="websocket-panel">
  //       Server Says: {serverMessage}
  //     </div>
  //
  //     {!playerSession ? (
  //       <button onClick={startGame}>Start Game</button>
  //     ) : (
  //       <div>
  //         <h3>{questions[currentQuestionIndex]?.questionText}</h3>
  //         {questions[currentQuestionIndex]?.options.map((option, index) => (
  //           <button
  //             key={index} // Key should be unique for each child in a list, use index as the key
  //             onClick={() => submitAnswer(index)} // Pass index instead of option
  //             className="option-button"
  //           >
  //             {option}
  //           </button>
  //         ))}
  //         <p>yay frontend</p>
  //         <p className="score">Score: {score}</p>
  //       </div>
  //     )}
  //   </div>
  // );

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      // Optionally, show a message confirming that the text was copied
      console.log("Lobby link copied to clipboard!");
    }, (err) => {
      console.error('Could not copy text: ', err);
    });
  };

  return (
      <div className="App">
        {!lobbySession ? (
            <div>
              {/* Section for creating a new lobby */}
              <button onClick={createNewLobby}>New Lobby</button>
            </div>
        ) : !gameStarted ? (
            <div>
              {/* Section for starting a game within an existing lobby */}
              <button onClick={startGame}>Start Game</button>
              <div>
                <p>Share this lobby link:</p>
                <input
                    type="text"
                    value={`${window.location.origin}/lobby/${lobbySession}`}
                    readOnly
                    onClick={(e) => {
                      e.target.select(); // Select the text to visually indicate that it's ready to be copied
                      copyToClipboard(e.target.value); // Call the function to copy the text
                    }}
                />
                <p>Click the link to copy and share it with others to join this lobby.</p>
              </div>
            </div>
        ) : (
            <div>
              {/* Game session UI */}
              <h3>{questions[currentQuestionIndex]?.questionText}</h3>
              {questions[currentQuestionIndex]?.options.map((option, index) => (
                  <button key={index} onClick={() => submitAnswer(index)} className="option-button">
                    {option}
                  </button>
              ))}
              <p className="score">Score: {score}</p>
            </div>
        )}
      </div>
  );
}

export default App;
