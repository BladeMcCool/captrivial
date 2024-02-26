import React, {createContext, useRef, useState} from "react";
import "./App.css";
import LobbyCreation from './components/lobbyCreation';
import StartGame from './components/startGame';
import GameOver from './components/gameOver';
import PickAnswer from "./components/pickAnswer";
import Countdown from "./components/countdown";
import useJoinLobby from './hooks/useJoinLobby';
import useWebsocketEventListener from "./hooks/useWebsocketEventListener";

// Use REACT_APP_BACKEND_URL or http://localhost:8080 as the API_BASE
const API_BASE = process.env.REACT_APP_BACKEND_URL || "http://localhost:8080";
export const AppContext = createContext();

function App() {
  const [lobbySession, setLobbySession] = useState(null);
  const [playerSession, setPlayerSession] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [currentQuestionAnswered, setCurrentQuestionAnswered] = useState(false);
  const [gameStarted, setGameStarted] = useState(false);
  const [gameEnded, setGameEnded] = useState(false);
  const [score, setScore] = useState(0);
  const [noPointsAwarded, setNoPointsAwarded] = useState(false);
  const [winnerMessage, setWinnerMessage] = useState(null);
  const [winningScore, setWinningScore] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [questionCount, setQuestionCount] = useState(3);
  const [countdownSeconds, setCountdownSeconds] = useState(5);
  const [countdownRunning, setCountdownRunning] = useState(false);
  const [countdownRemainingMs, setCountdownRemainingMs] = useState(0);
  const hasJoinedLobby = useRef(false); // using this to very aggressively prevent double execution of lobby-joining since the server is responsible for generating and adding the new session, doing it more than once is bad.

  useWebsocketEventListener(API_BASE, playerSession, lobbySession, setGameStarted, setCountdownRunning, setCountdownRemainingMs, setQuestions, setGameEnded, setWinnerMessage, setWinningScore, setError, setLoading, setNoPointsAwarded, setCurrentQuestionAnswered)
  useJoinLobby(API_BASE, setLobbySession, setPlayerSession, setError, setLoading, hasJoinedLobby);

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
            />
        ) : countdownRunning ? (
            <Countdown
                countdownRemainingMs={countdownRemainingMs}
                setCountdownRemainingMs={setCountdownRemainingMs}
                countdownRunning={countdownRunning}
                setCountdownRunning={setCountdownRunning}
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
            />
        ) : noPointsAwarded && currentQuestionAnswered ? (
            <div>
              <h2>Incorrect Response, Sorry!</h2>
              <p>Please wait while the next question is being transferred</p>
            </div>
        ) : (
            <PickAnswer
                questions={questions}
                score={score}
                lobbySession={lobbySession}
                playerSession={playerSession}
                setScore={setScore}
                setError={setError}
                setLoading={setLoading}
                setNoPointsAwarded={setNoPointsAwarded}
                setCurrentQuestionAnswered={setCurrentQuestionAnswered}
            />
        )}
      </div>
      </AppContext.Provider>
  );
}

export default App;
