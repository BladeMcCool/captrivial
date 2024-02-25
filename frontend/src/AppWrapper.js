import {BrowserRouter as Router, Route, Routes} from "react-router-dom";
import React from "react";
import App from "./App";

function AppWrapper() {
    return (
        <Router>
            <Routes>
                <Route path="/lobby/:lobbyUuid" element={<App/>}/>
                <Route path="/" element={<App/>}/>
            </Routes>
        </Router>
    );
}

export default AppWrapper; // Export AppWrapper for use in index.js