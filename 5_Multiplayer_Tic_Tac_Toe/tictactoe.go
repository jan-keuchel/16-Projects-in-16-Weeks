package main

import (
	"strconv"
	"strings"
)

type TTT struct {
	gameBoard  	    [9]string
	numOfValidMoves int
}

// Returns a pointer to a struct of a Tic-Tac-Toe game which is initialized to an
// empty board and a start player of "X"
func NewTTT() *TTT {
	return &TTT{
		gameBoard:  	 [9]string{"_", "_", "_", "_", "_", "_", "_", "_", "_"},
		numOfValidMoves: 0,
	}
}

// Resets the state of the game.
func (t *TTT) reset() {

	t.gameBoard       = [9]string{"_", "_", "_", "_", "_", "_", "_", "_", "_"}
	t.numOfValidMoves = 0

}

// Board:
//   0 1 2\n
// 0 _ _ _\n
// 3 _ _ _\n
// 6 _ _ _\n

// printBoard returns a string which represents the current state of the game.
func (t *TTT) printBoard() string {
	
	var builder strings.Builder
	builder.WriteString( "  0 1 2\n")
	for i := range 3 {
		builder.WriteString(strconv.Itoa(i * 3))
		for j := 3 * i; j < 3 * (i + 1); j++ {
			builder.WriteString(" " + t.gameBoard[j])
		}
		builder.WriteString("\n")
	}
	return builder.String()

}

// Checks if a move can be made (not already taken), makes that move for the 
// according player and increments the total moves made in the game. 
// Returns true for the first argument if move was made successfully and false if not.
// Returns true for the second argument if there is a tie (9 Moves made)
func (t *TTT) stepGame(cell int, player string) (bool, bool) {

	if t.gameBoard[cell] != "_" {
		return false, false
	} 
	t.gameBoard[cell] = player
	t.numOfValidMoves++
	if t.numOfValidMoves < 9 {
		return true, false
	}
	return true, true

}

// checkForWinner tests whether the player who made the last move won the game.
// Returns true if he did, false otherwise.
func (t *TTT) checkForWinner(cell int, player string) bool {

	// horizontal check
	offset := cell / 3
	foundWinner := true
	for i := range 3 {
		if t.gameBoard[3 * offset + i] != player {
			foundWinner = false
			break
		}
	}
	if foundWinner {
		return true
	}

	// vertical check
	offset = cell % 3
	foundWinner = true
	for i := range 3 {
		if t.gameBoard[offset + 3 * i] != player {
			foundWinner = false
			break
		}
	}
	if foundWinner {
		return true
	}

	// diagonal checks
	// bottom left or top right corner (Ascending diagonal)
	// indices 2, 4, 6
	foundWinner = true
	if cell == 6 || cell == 2 || cell == 4 {
		base := 2
		for i := range 3 {
			if t.gameBoard[base * 2 * i] != player {
				foundWinner = false
				break
			}
		}
		if foundWinner {
			return true
		}
	}

	// top left or bottom right corner (Descending diagonal)
	// indices 0, 4, 8
	foundWinner = true
	if cell == 0 || cell == 8 || cell == 4 {
		for i := range 3 {
			if t.gameBoard[4 * i] != player {
				foundWinner = false
				break
			}
		}
		if foundWinner {
			return true
		}
	}

	return false

}
