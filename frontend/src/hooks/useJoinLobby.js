import { useEffect } from 'react';
import { useParams } from 'react-router-dom';

const useJoinLobby = (API_BASE, setLobbySession, setPlayerSession, setError, setLoading, hasJoinedLobby) => {

    // Use useParams hook to extract lobby UUID from the URL
    const { lobbyUuid } = useParams(); // Extract lobbyUuid from URL

    useEffect(() => {
        if (lobbyUuid) {
            setLobbySession(lobbyUuid);
            console.log("there is a lobbyUuid and it is", lobbyUuid)
            if (!hasJoinedLobby.current) { //react strict mode was causing this to double-join lobbies and since server sets the session id we were having 2 session ids join the game, and the '3rd player walks away from the computer' issue arose. trying to use state to track the fact we got a session was not registering fast enough so this is an alternate method for immediate 'state' (not actual react state i guess?) adjustment.
                hasJoinedLobby.current = true
                // Function to join the lobby if we haven't got a playerSession yet
                const joinLobby = async () => {
                    setLoading(true);
                    try {
                        const response = await fetch(`${API_BASE}/game/joinlobby/${lobbyUuid}`, {
                            method: "GET",
                            headers: {
                                "Content-Type": "application/json",
                            },
                        });

                        if (!response.ok) {
                            throw new Error("Failed to join lobby");
                        }
                        const data = await response.json();
                        // Handle successful lobby join
                        console.log("got data from attempt to joinlobby", data)
                        setPlayerSession(data.sessionId)
                        console.log("player joined game with session id", data.sessionId)
                    } catch (error) {
                        setError(error.message);
                    } finally {
                        setLoading(false);
                    }
                };

                console.log("there is not a playersession ... (yet)")
                joinLobby();
            }
        }
    }, [lobbyUuid, setLobbySession, setPlayerSession, setError, setLoading, API_BASE, hasJoinedLobby]);
};

export default useJoinLobby;
