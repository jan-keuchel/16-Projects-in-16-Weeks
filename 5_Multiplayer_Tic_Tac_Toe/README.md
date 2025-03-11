# Multiplayer Tic-tac-toe

## Goal
The goal is to build a terminal based multiplayer Tic-tac-toe game where one can start a server and then connect two clients to it via TCP.

## Features
- [x] Single file with control via terminal interaction (Chooese between server and client)
    - [x] "Welcome to Tic-Tac-Toe, [S]erver, [C]lient:"
    - [x] (S) -> "Port:"
    - [x] (C) -> "'IP:Port':"
- [x] Two players allowed:
    - [x] As long as one player is present the game doesn't start. If another player joins, the game begins.
    - [x] First player is X, second player is O
    - [x] Field selection via numbers from 0 to 8
        - [x] Logging tells where the other player placed his move
        - [x] Terminal graphic representation of the game board
    - [x] Automatic validation of winner with winner- and loser-message
    - [x] Rematch option
- [ ] Proper connection handling
    - [ ] Players get notified if opponent disconnects
    - [ ] Client gets notified and exits cleanly if server shuts down
