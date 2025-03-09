package main

type TTT struct {
	gameBoard  	  [9]string
	currentPlayer string
}

func NewTTT() *TTT {
	return &TTT{
		gameBoard:     [9]string{" ", " ", " ", " ", " ", " ", " ", " ", " "},
		currentPlayer: "X",
	}
}
