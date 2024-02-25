import React from 'react';

const LobbyCreation = ({ questionCount, setQuestionCount, countdownSeconds, setCountdownSeconds, createNewLobby }) => {
    return (
        <div className="create-lobby-container">
            <h2>Create New Lobby</h2>
            <div className="lobby-settings">
                <div className="setting">
                    <label>
                        Number of Questions:
                        <select value={questionCount} onChange={(e) => setQuestionCount(Number(e.target.value))}>
                            {[1, 2, 3, 5, 10, 20].map(num => (
                                <option key={num} value={num}>{num}</option>
                            ))}
                        </select>
                    </label>
                </div>
                <div className="setting">
                    <label>
                        Countdown Seconds:
                        <select value={countdownSeconds} onChange={(e) => setCountdownSeconds(Number(e.target.value))}>
                            {[1, 3, 5, 10].map(sec => (
                                <option key={sec} value={sec}>{sec}</option>
                            ))}
                        </select>
                    </label>
                </div>
            </div>
            <div className="start-button-container">
                <button onClick={createNewLobby}>Create New Lobby</button>
            </div>
        </div>
    );
};

export default LobbyCreation;