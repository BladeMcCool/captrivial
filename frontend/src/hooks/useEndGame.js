const useEndGame = () => {
    const endGame = async (API_BASE, lobbySession, playerSession, setGameEnded, setWinnerMessage, setWinningScore, setError, setLoading) => {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE}/game/status/${lobbySession}`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                },
            });
            const data = await res.json();
            // alert(`Game over! Your score: ${data.finalScore}`); // Use the finalScore from the response
            console.log("game status after end: ", data)

            // Check if the current user is a winner
            const isWinner = data.winners.includes(playerSession);
            const you = "You"

            // Replace the current user's session ID with "you" and reorder to put "you" first if present
            const winnersFormatted = data.winners.map(winner => winner === playerSession ? you : winner);
            if (isWinner && winnersFormatted.length > 1) {
                const index = winnersFormatted.indexOf(you);
                winnersFormatted.splice(index, 1); // Remove "you"
                winnersFormatted.unshift(you); // Add "you" to the beginning
            }

            const winnerMessage = isWinner ?
                `Congratulations, ${winnersFormatted.join(" and ")} won the game!` :
                `${winnersFormatted.join(" and ")} won the game!`;

            setWinnerMessage(winnerMessage)
            setWinningScore(data.winningScore);
            setGameEnded(true)
        } catch (err) {
            setError("Failed to end game.");
        }
        setLoading(false);
    }
    return endGame;
}

export default useEndGame;