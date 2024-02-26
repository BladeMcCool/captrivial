import { useEffect } from 'react';
import useEndGame from "./useEndGame";

const useWebsocketEventListener = (API_BASE, playerSession, lobbySession, setGameStarted, setCountdownRunning, setCountdownRemainingMs, setQuestions, setGameEnded, setWinnerMessage, setWinningScore, setError, setLoading, setNoPointsAwarded, setCurrentQuestionAnswered) => {
    // Effect for WebSocket setup
    const endGame = useEndGame();

    useEffect(() => {
        if (!playerSession || !lobbySession) return; // Only connect WebSocket after lobby is waiting
        const API_WS = API_BASE.replace(/^http/, "ws");
        const websocket = new WebSocket(`${API_WS}/game/events/${lobbySession}/${playerSession}`);

        websocket.onopen = () => {
            console.log('WebSocket Connected');
        };

        websocket.onmessage = async (event) => {
            // setServerMessage(event.data);
            const data = JSON.parse(event.data);
            console.log("got message from server", data)
            // Handle different types of messages
            if (data.question) {
                console.log("received question")
                setGameStarted(true)
                setCountdownRunning(false)
                setCountdownRemainingMs(0)
                setNoPointsAwarded(false)
                setQuestions(prev => [...prev, data.question]);
                console.log("saying we have not answered the current question")
                setCurrentQuestionAnswered(false)
            } else if (data.countdownMs) {
                console.log("show countdown ticker for this many ms:", data.countdownMs)
                setCountdownRunning(true)
                setCountdownRemainingMs(data.countdownMs)
            } else if (data.gameOver) {
                try {
                    //switch to the gameOver screen with winner info.
                    await endGame(API_BASE, lobbySession, playerSession, setGameEnded, setWinnerMessage, setWinningScore, setError, setLoading)
                } catch (error) {
                    console.error("Failed to end the game:", error);
                }
            } else if (data.content) {
                //websocket debugging
                console.log("spam data", data)
            }
        };

        return () => {
            console.log('Closing WebSocket...');
            websocket.close();
        };
    // TODO learn about the useCallback stuff that could maybe solve this lint
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [playerSession, lobbySession]); // Re-connect WebSocket if playerSession or lobby changes

}
export default useWebsocketEventListener;
