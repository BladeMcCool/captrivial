import React from 'react';

const Countdown = ({ countdownRemainingMs }) => {
    return (
        <div>
            <h2>Get Ready!</h2>
            <p>Game starts in {(countdownRemainingMs / 1000).toFixed(1)} seconds</p>
        </div>
    );
};

export default Countdown;