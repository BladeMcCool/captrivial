import React from 'react';

const GameOver = ({ winnerMessage, score, winningScore, resetGame }) => {
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