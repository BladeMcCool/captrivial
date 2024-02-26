import React, {useEffect} from 'react';


const Countdown = ({ countdownRemainingMs, countdownRunning, setCountdownRemainingMs, setCountdownRunning }) => {
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
    }, [countdownRunning, countdownRemainingMs, setCountdownRemainingMs, setCountdownRunning ]);

    return (
        <div>
            <h2>Get Ready!</h2>
            <p>Game starts in {(countdownRemainingMs / 1000).toFixed(1)} seconds</p>
        </div>
    );
};

export default Countdown;