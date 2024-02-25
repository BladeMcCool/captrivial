import React, {useContext} from 'react';
import {AppContext} from "../App";

const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(() => {
        // Optionally, show a message confirming that the text was copied
        console.log("Lobby link copied to clipboard!");
    }, (err) => {
        console.error('Could not copy text: ', err);
    });
};

const StartGame = ({ lobbySession, playerSession, setLoading, setError }) => {
    const { API_BASE } = useContext(AppContext);

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
        } catch (err) {
            setError("Failed to start game.");
        }
        setLoading(false);
    };

    return (
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
    );
};

export default StartGame;