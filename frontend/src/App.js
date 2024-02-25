import React, {createContext, useEffect, useRef, useState} from "react";
import "./App.css";
import { useParams } from "react-router-dom"; // Import useParams hook
import LobbyCreation from './components/lobbyCreation';
import StartGame from './components/startGame';
import GameOver from './components/gameOver';
import PickAnswer from "./components/pickAnswer";
import Countdown from "./components/countdown";
import useEndGame from './hooks/useEndGame';
import useJoinLobby from './hooks/useJoinLobby';

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
  const endGame = useEndGame();

  // Effect for WebSocket setup
  useEffect(() => {
    if (!playerSession || !lobbySession) return; // Only connect WebSocket after lobby is waiting

    const websocket = new WebSocket(`ws://localhost:8080/game/events/${lobbySession}/${playerSession}`);

    websocket.onopen = () => {
      console.log('WebSocket Connected');
    };

    websocket.onmessage = async (event) => {
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
        // endGame()
        try {
          await endGame(API_BASE, lobbySession, playerSession, setGameEnded, setWinnerMessage, setWinningScore, setError, setLoading)
          // Handle success
        } catch (error) {
          // Handle errors
          console.error("Failed to end the game:", error);
        }

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

  const hasJoinedLobby = useRef(false);

  useJoinLobby(API_BASE, setLobbySession, setPlayerSession, setError, setLoading, hasJoinedLobby);


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
