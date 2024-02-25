import React, {useContext} from 'react';
import { AppContext } from '../App';


const LobbyCreation = ({ questionCount, setQuestionCount, countdownSeconds, setCountdownSeconds, setPlayerSession, setLobbySession, setError, setLoading }) => {
    const { API_BASE } = useContext(AppContext);

    const createNewLobby = async () => {
        try {
            try {
                const res = await fetch(`${API_BASE}/game/newlobby`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({
                        questionCount: questionCount,
                        countdownMs: countdownSeconds * 1000, //just using seconds for the ui
                    }),
                });

                const data = await res.json();
                if (res.ok) {
                    console.log("game lobby:", data)
                    // Assuming the response includes the lobbySession or playerSession identifier
                    // setLobbySession(data.lobbyId); // Update this line based on your actual response structure
                    setPlayerSession(data.sessionId);
                    setLobbySession(data.lobbyId);
                    // setGameParams(data);
                    // Additional logic to handle successful lobby creation
                } else {
                    throw new Error(data.error || "Failed to create new lobby");
                }
            } catch (err) {
                setError(err.message);
            }
        } catch (err) {
            setError("Failed to create lobby");
        }
        setLoading(false);
    };

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