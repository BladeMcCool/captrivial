import React from 'react';

const PickAnswer = ({ questions, score, submitAnswer }) => {
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