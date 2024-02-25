import React, {createContext, useEffect, useRef, useState} from "react";
import "./App.css";
import { useParams } from "react-router-dom"; // Import useParams hook
import LobbyCreation from './components/lobbyCreation';
import StartGame from './components/startGame';
import GameOver from './components/gameOver';
import PickAnswer from "./components/pickAnswer";
import Countdown from "./components/countdown";

// Use REACT_APP_BACKEND_URL or http://localhost:8080 as the API_BASE
const API_BASE = process.env.REACT_APP_BACKEND_URL || "http://localhost:8080";
export const AppContext = createContext();

function App() {
  const [lobbySession, setLobbySession] = useState(null);
  const [playerSession, setPlayerSession] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [gameStarted, setGameStarted] = useState(false);
  const [gameEnded, setGameEnded] = useState(false);
  const [score, setScore] = useState(0);
  const [winnerMessage, setWinnerMessage] = useState(null);
  const [winningScore, setWinningScore] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [questionCount, setQuestionCount] = useState(3);
  const [countdownSeconds, setCountdownSeconds] = useState(5);
  const [countdownRunning, setCountdownRunning] = useState(false);
  const [countdownRemainingMs, setCountdownRemainingMs] = useState(0);

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
        setCountdownRunning(false)
        setCountdownRemainingMs(0)
        setQuestions(prev => [...prev, data.question]);
        // setCurrentQuestionIndex(prevIndex => prevIndex + 1); // Move to the new question
      } else if (data.countdownMs) {
        console.log("maybe show some kind of ticker for this many ms:", data.countdownMs)
        setCountdownRunning(true)
        setCountdownRemainingMs(data.countdownMs)
      } else if (data.gameOver) {
        endGame()
        // alert(`Game over! Winner: ${data.winner}, Score: ${data.score}`);
        // Reset game state here if needed
      } else if (data.content) {
        //websocket debugging
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
  const hasJoinedLobby = useRef(false);

  // Effect to set lobbySession based on URL lobbyUuid
  useEffect(() => {
    if (lobbyUuid) {
      setLobbySession(lobbyUuid);
      console.log("there is a lobbyUuid and it is", lobbyUuid)
      if (!hasJoinedLobby.current) { //react strict mode was causing this to double-join lobbies and since server sets the session id we were having 2 session ids join the game, and the '3rd player walks away from the computer' issue arose. trying to use state to track the fact we got a session was not registering fast enough so this is an alternate method for immediate 'state' (not actual react state i guess?) adjustment.
        hasJoinedLobby.current = true
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

  useEffect(() => {
    let intervalId;
    if (countdownRunning && countdownRemainingMs > 0) {
      //TODO better time calc logic based on taking the actual time the game was started at, and showing how much time is left between now and game start.
      //but this should be good enough for a first draft.
      intervalId = setInterval(() => {
        setCountdownRemainingMs((time) => time - 100);
      }, 100);
    } else if (countdownRemainingMs <= 0) {
      setCountdownRunning(false);
    }

    return () => clearInterval(intervalId); // Cleanup interval on unmount or when countdownRemainingMs becomes 0 or less
  }, [countdownRunning]);

  // const createNewLobby = async () => {
  //   try {
  //     try {
  //       const res = await fetch(`${API_BASE}/game/newlobby`, {
  //         method: "POST",
  //         headers: {
  //           "Content-Type": "application/json",
  //         },
  //         body: JSON.stringify({
  //           questionCount: questionCount,
  //           countdownMs: countdownSeconds * 1000, //just using seconds for the ui
  //         }),
  //       });
  //
  //       const data = await res.json();
  //       if (res.ok) {
  //         console.log("game lobby:", data)
  //         // Assuming the response includes the lobbySession or playerSession identifier
  //         // setLobbySession(data.lobbyId); // Update this line based on your actual response structure
  //         setPlayerSession(data.sessionId);
  //         setLobbySession(data.lobbyId);
  //         // setGameParams(data);
  //         // Additional logic to handle successful lobby creation
  //       } else {
  //         throw new Error(data.error || "Failed to create new lobby");
  //       }
  //     } catch (err) {
  //       setError(err.message);
  //     }
  //   } catch (err) {
  //     setError("Failed to create lobby");
  //   }
  //   setLoading(false);
  // };

  // const startGame = async () => {
  //   setLoading(true);
  //   setError(null);
  //   try {
  //     const res = await fetch(`${API_BASE}/game/start`, {
  //       method: "POST",
  //       headers: {
  //         "Content-Type": "application/json",
  //       },
  //       body: JSON.stringify({
  //         lobbyId: lobbySession,
  //         sessionId: playerSession,
  //       }),
  //     });
  //     const data = await res.json();
  //     console.log("got data from start game:", data)
  //     // setPlayerSession(data.sessionId);
  //     // fetchQuestions();
  //   } catch (err) {
  //     setError("Failed to start game.");
  //   }
  //   setLoading(false);
  // };

  // const fetchQuestions = async () => {
  //   setLoading(true);
  //   try {
  //     const res = await fetch(`${API_BASE}/questions`);
  //     const data = await res.json();
  //     setQuestions(data);
  //   } catch (err) {
  //     setError("Failed to fetch questions.");
  //   }
  //   setLoading(false);
  // };

  // const submitAnswer = async (index) => {
  //   // We are submitting the index
  //   setLoading(true);
  //   const currentQuestion = questions[questions.length-1];
  //   console.log("believes current question id is: ", currentQuestion.id)
  //   try {
  //     const res = await fetch(`${API_BASE}/game/answer`, {
  //       method: "POST",
  //       headers: {
  //         "Content-Type": "application/json",
  //       },
  //       body: JSON.stringify({
  //         lobbyId: lobbySession,
  //         sessionId: playerSession,
  //         questionId: currentQuestion.id, // field name is "id", not "questionId"
  //         answer: index,
  //       }),
  //     });
  //     const data = await res.json();
  //     if (data.points) {
  //       //TODO points will tell us if we got the question right or not ... do something fancy if so
  //       setScore(data.score); // Update score from server's response
  //     }
  //   } catch (err) {
  //     setError("Failed to submit answer.");
  //   }
  //   setLoading(false);
  // };

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

  // const resetGame = async () => {
  //   setLoading(true);
  //   // change something to go back to the "new lobby screen"
  //   setError(null);
  //   setPlayerSession(null);
  //   setLobbySession(null);
  //   hasJoinedLobby.current = false
  //   setQuestions([]);
  //   setScore(0);
  //   setGameStarted(false)
  //   setGameEnded(false)
  //   setLoading(false);
  // };

  if (error) return <div className="error">Error: {error}</div>;
  if (loading) return <div className="loading">Loading...</div>;

  return (
      <AppContext.Provider value={{ API_BASE }}>
      <div className="App">
        {!lobbySession ? (
            <LobbyCreation
                questionCount={questionCount}
                setQuestionCount={setQuestionCount}
                countdownSeconds={countdownSeconds}
                setCountdownSeconds={setCountdownSeconds}
                setPlayerSession={setPlayerSession}
                setLobbySession={setLobbySession}
                setError={setError}
                setLoading={setLoading}
                // createNewLobby={createNewLobby}
            />
        ) : countdownRunning ? (
            <Countdown
                countdownRemainingMs={countdownRemainingMs}
            />
        ) : !gameStarted ? (
            <StartGame
                lobbySession={lobbySession}
                playerSession={playerSession}
                setLoading={setLoading}
                setError={setError}
            />
        ) : gameEnded ? (
            <GameOver
                winnerMessage={winnerMessage}
                score={score}
                winningScore={winningScore}
                setLoading={setLoading}
                setError={setError}
                setPlayerSession={setPlayerSession}
                setLobbySession={setLobbySession}
                hasJoinedLobby={hasJoinedLobby}
                setQuestions={setQuestions}
                setScore={setScore}
                setGameStarted={setGameStarted}
                setGameEnded={setGameEnded}
                // resetGame={resetGame}
            />
        ) : (
            <PickAnswer
                questions={questions}
                score={score}
                lobbySession={lobbySession}
                playerSession={playerSession}
                setScore={setScore}
                setError={setError}
                setLoading={setLoading}
            />
        )}
      </div>
      </AppContext.Provider>
  );
}

export default App;
