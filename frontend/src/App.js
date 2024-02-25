import React, { useEffect, useState } from "react";
import "./App.css";
import { useParams } from "react-router-dom"; // Import useParams hook

// Use REACT_APP_BACKEND_URL or http://localhost:8080 as the API_BASE
const API_BASE = process.env.REACT_APP_BACKEND_URL || "http://localhost:8080";

function App() {
  // TODO make these inputs on the screen, with a minimum countdown MS like 3000
  const countDownMs = 1000
  const questionCount = 2

  const [lobbySession, setLobbySession] = useState(null);
  const [playerSession, setPlayerSession] = useState(null);
  const [questions, setQuestions] = useState([]);
  // const [gameParams, setGameParams] = useState({});
  const [gameStarted, setGameStarted] = useState(false);
  const [gameEnded, setGameEnded] = useState(false);
  // const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [score, setScore] = useState(0);
  const [winnerMessage, setWinnerMessage] = useState(null);
  const [winningScore, setWinningScore] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  // const [ws, setWs] = useState(null); // Store WebSocket connection
  // const [serverMessage, setServerMessage] = useState('');

  // Effect for WebSocket setup
  useEffect(() => {
    if (!playerSession || !lobbySession) return; // Only connect WebSocket after lobby is waiting

    const websocket = new WebSocket(`ws://localhost:8080/game/events/${lobbySession}/${playerSession}`);

    websocket.onopen = () => {
      console.log('WebSocket Connected');
    };

    websocket.onmessage = (event) => {
      // setServerMessage(event.data);
      const data = JSON.parse(event.data);
      console.log("got message from server", data)
      // Handle different types of messages
      if (data.question) {
        console.log("received question")
        setGameStarted(true)
        setQuestions(prev => [...prev, data.question]);
        // setCurrentQuestionIndex(prevIndex => prevIndex + 1); // Move to the new question
      } else if (data.countdownMs) {
        console.log("maybe show some kind of ticker for this many ms:", data.countdownMs)
      } else if (data.gameOver) {
        endGame()
        // alert(`Game over! Winner: ${data.winner}, Score: ${data.score}`);
        // Reset game state here if needed
      } else if (data.content) {
        console.log("spam data", data)
        // Reset game state here if needed
      }
      // Add more conditions as needed based on your server messages
    };

    // setWs(websocket); // Store WebSocket connection

    return () => {
      console.log('Closing WebSocket...');
      websocket.close();
    };
  }, [playerSession, lobbySession]); // Re-connect WebSocket if playerSession changes

  // Use useParams hook to extract lobby UUID from the URL
  let { lobbyUuid } = useParams();
  // console.log("hrm")

  // Effect to set lobbySession based on URL lobbyUuid
  useEffect(() => {
    if (lobbyUuid) {
      setLobbySession(lobbyUuid);
      console.log("there is a lobbyUuid and it is", lobbyUuid)
      if (!playerSession) {
        // Function to join the lobby if we haven't got a playerSession yet
        const joinLobby = async () => {
          setLoading(true);
          try {
            const response = await fetch(`${API_BASE}/game/joinlobby/${lobbyUuid}`, {
              method: "GET",
              headers: {
                "Content-Type": "application/json",
              },
            });

            if (!response.ok) {
              throw new Error("Failed to join lobby");
            }
            const data = await response.json();
            // Handle successful lobby join
            console.log("got data from attempt to joinlobby", data)
            setPlayerSession(data.sessionId)
            console.log("player joined game with session id", data.sessionId)
          } catch (error) {
            setError(error.message);
          } finally {
            setLoading(false);
          }
        };

        console.log("there is not a playersession ... (yet)")
        joinLobby();
      }
    }
  }, [lobbyUuid]);

  const createNewLobby = async () => {
    try {
      try {
        const res = await fetch(`${API_BASE}/game/newlobby`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            questionCount: questionCount,
            countdownMs: countDownMs,
          }),
        });
        const data = await res.json();
        if (res.ok) {
          console.log("game lobby:", data)
          // Assuming the response includes the lobbySession or playerSession identifier
          // setLobbySession(data.lobbyId); // Update this line based on your actual response structure
          setPlayerSession(data.sessionId);
          setLobbySession(data.lobbyId);
          // setGameParams(data);
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
        body: JSON.stringify({
          lobbyId: lobbySession,
          sessionId: playerSession,
        }),
      });
      const data = await res.json();
      console.log("got data from start game:", data)
      // setPlayerSession(data.sessionId);
      // fetchQuestions();
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

  const endGame = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/game/status/${lobbySession}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });
      const data = await res.json();
      // alert(`Game over! Your score: ${data.finalScore}`); // Use the finalScore from the response
      console.log("game status after end: ", data)

      // Check if the current user is a winner
      const isWinner = data.winners.includes(playerSession);
      const you = "You"

      // Replace the current user's session ID with "you" and reorder to put "you" first if present
      const winnersFormatted = data.winners.map(winner => winner === playerSession ? you : winner);
      if (isWinner && winnersFormatted.length > 1) {
        const index = winnersFormatted.indexOf(you);
        winnersFormatted.splice(index, 1); // Remove "you"
        winnersFormatted.unshift(you); // Add "you" to the beginning
      }

      const winnerMessage = isWinner ?
          `Congratulations, ${winnersFormatted.join(" and ")} won the game!` :
          `${winnersFormatted.join(" and ")} won the game!`;

      setWinnerMessage(winnerMessage)
      // setWinners(data.winners);
      setWinningScore(data.winningScore);

      // TODO do all this stuff before starting a lobby, not here.
      // setPlayerSession(null);
      // setLobbySession(null);
      // setQuestions([]);
      // setScore(0);

      // setCurrentQuestionIndex(0);
      setGameEnded(true)
    } catch (err) {
      setError("Failed to end game.");
    }
    setLoading(false);
  };

  const resetGame = async () => {
    setLoading(true);
    // change something to go back to the "new lobby screen"
    setError(null);
    setPlayerSession(null);
    setLobbySession(null);
    setQuestions([]);
    setScore(0);
    setGameStarted(false)
    setGameEnded(false)
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
                    className="lobby-link-input"
                />
                <p>Click the link to copy and share it with others to join this lobby.</p>
              </div>
            </div>
        ) : gameEnded ? (
            <div>
              <div>
                <h2>Game Over</h2>
                <p>{winnerMessage}</p>
                <p>Your Score: {score}</p>
                <p>Winning Score: {winningScore}</p>
                <button onClick={resetGame}>Reset Game</button>
              </div>
            </div>
        ) : (
            <div>
            {/* Game session UI */}
              <h3>{questions[questions.length - 1]?.questionText}</h3>
              {questions[questions.length-1]?.options.map((option, index) => (
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
