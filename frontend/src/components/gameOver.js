import React from 'react';

const GameOver = ({ winnerMessage, score, winningScore, setLoading, setError, setPlayerSession, setLobbySession, hasJoinedLobby, setQuestions, setScore, setGameStarted, setGameEnded }) => {
    const resetGame = async () => {
        setLoading(true);
        // change something to go back to the "new lobby screen"
        setError(null);
        setPlayerSession(null);
        setLobbySession(null);
        hasJoinedLobby.current = false
        setQuestions([]);
        setScore(0);
        setGameStarted(false)
        setGameEnded(false)
        setLoading(false);
    };

    return (
        <div>
            <div>
                <h2>Game Over</h2>
                <p>{winnerMessage}</p>
                <p>Your Score: {score}</p>
                <p>Winning Score: {winningScore}</p>
                <button onClick={resetGame}>Reset Game</button>
            </div>
        </div>
    );
};

export default GameOver;