# CapTrivia: Cap Table Trivia Game

### Pulley take home assessment - 2024-02-25 (Feb 25 2024)
I spent about 3 days to take it from its initial basic state and transformed it into a real time multiplayer trivia game where players could compete to be the first to answer questions. A websocket to the connected clients pushed the updates down to a React front end. Players could join a game via a shareable link. The project was dockerized.

***

## Introduction
Welcome to CapTrivia, a cap table trivia game developed by Pulley. This project is part of our engineering interview process, designed to assess your coding skills and problem-solving abilities. Currently, CapTrivia is a simple, single-player game. Your challenge is to develop it into a multiplayer experience.

## Project Overview
- **Current State**: CapTrivia is a functioning single-player game with basic features. We tried our best to make it very ugly so that you have lots of room to improve it!
- **Tech Stack**: The front end is built with React (in `App.js`), and the back end with Go (in `main.go`). We also included a docker-compose.yml so that you can easily run this locally (it includes a PostgreSQL instance).
- **Objective**: Transform CapTrivia into a multiplayer game where users can create games, invite others via a shareable link, and compete in real-time trivia challenges.

## Quickstart

1. Make sure you have Docker installed. https://docs.docker.com/engine/install/
2. In the captrivia root directory, run `docker compose up`.
3. Open http://localhost:3000 in your browser to see the game.

## Tasks
Please complete the task appropriate to the position you are applying for.

### Software Engineers
1. **Refactor and Enhance**: The current code is intentionally basic and contained in single files for both front and back ends. Your first task is to clean up and structure the codebase.
2. **Implement Multiplayer Functionality**: 
    - Allow users to start a new game and generate a shareable link.
    - Enable real-time joining of players in a waiting room.
    - Implement a countdown and start the game simultaneously for all players.
    - Introduce a scoring system with a configurable number of questions (default: 10).
    - Players will compete to answer each question first. For each question, the player who answers first gets the points.
    - Display the winner at the end and allow users to start new games.

### Data Engineers
Your task is to build a data pipeline to record and visualize analytics information about a player's performance.
  - Stand up an analytics database: you may use the existing PostgreSQL database in the docker-compose.yml, or any other database of your choice.
  - Instrument the backend application to report data on each answered question to the database.
  - Start up the bot provided in [bot](bot) to start playing games and feeding data into your database.  Right now the bot uses a very simple strategy (always choosing the first answer); you will likely want to change its strategy to create more interesting data.  If you like, your bot can "cheat" by using [questions.json](backend/questions.json) to sometimes figure out the right answer.
  - Build and display some simple analytics on top of your analytics database.


## Time Allocation
There is no time restrictions on this challenge, you can take as much time as you need. This is to allow folks who might not know the tech stack to spend time learning or if you wanted to take the challenge to the next level. The most important part is achieving the task above.

As a baseline for someone proficient in the stack we expect this challenge to take around 3-5 hours. This figure is only to help set your own expectations about how much time you may need to allocate to complete it.

## Submission Guidelines
- **Zip File**: Zip up the code directory and send it to us. Please make sure you delete `node_modules` and any Go binaries that would bloat the zip. Leaving the `.git` directory is preferred so we can see the history of how you implemented your solution. To automatically clean files that are not in git, you can use `git clean -fdx` (with appropriate caution as this deletes files).
- **Running Locally**: Please ensure that we can run your version with `docker compose up`.

## Evaluation Criteria
- **Code Quality**: Clean, readable, and well-structured code.
- **Problem Solving**: Effective and efficient solutions to the challenges presented.
- **Creativity**: We appreciate innovative approaches and ideas.
- **Tool Utilization**: Feel free to use tools like GitHub Copilot and ChatGPT. We value efficiency and resourcefulness.

## After Submission
Once we receive your solution, we'll schedule a live code pairing session. Here, you'll walk us through your code, and we'll collaboratively work on an additional small feature.

## Conclusion
Thank you for participating in Pulley's engineering challenge. We're excited to see your approach to enhancing CapTrivia. Good luck, and have fun!
