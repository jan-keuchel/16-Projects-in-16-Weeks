# Multiplayer Tic-tac-toe

## Goal
The goal is to build a terminal based multiplayer Tic-tac-toe game where one can start a server and then connect two clients to it via TCP.

## Features
- [ ] Single file with control via terminal interaction (Chooese between server and client)
    - [ ] "Welcome to Tic-Tac-Toe, [S]erver, [C]lient:"
    - [ ] (H) -> "Port:"
    - [ ] (C) -> "'IP:Port':"
- [ ] Two players allowed:
    - [ ] As long as one player is present the game doesn't start. If another player joins, the game begins.
    - [ ] First player is X, second player is O
    - [ ] Field selection via numbers from 0 to 8
        - [ ] Logging tells where the other player placed his move
        - [ ] Terminal graphic representation of the game board
    - [ ] Automatic validation of winner with winner- and loser-message
    - [ ] Rematch option
