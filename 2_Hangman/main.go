package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode/utf8"
)

func main() {

	var remaining_tries uint8 = 11
	var target string

	fmt.Println("Welcome to 'Hangman'. You'll need 2 players. The first player chooses a word which the second player has to guess. The second player will have 11 wrong guesses.")
	fmt.Println("Player 1, choose a word:")

	fmt.Scan(&target)
	fmt.Println("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\nPlayer 1 has successfully chosen a word.")

	var target_len = utf8.RuneCountInString(target)
	var correctly_guessed_runes = make([]bool, target_len)
	var number_of_correct_guesses uint8 = 0

	var reader = bufio.NewReader(os.Stdin)
	var newline_input bool = false

	for int(number_of_correct_guesses) < target_len  && remaining_tries > 0 {

		if !newline_input {
			fmt.Print("progress: ")
			for i, r := range target {
				if correctly_guessed_runes[i] {
					fmt.Printf("%c ", r)
				} else {
					fmt.Printf("_ ")
				}
			}
			fmt.Println("Remaining tries:", remaining_tries, "Player 2, guess a character:")
		}
		guess, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println("Error while reading guess.")
			return
		} else if guess == '\n' {
			newline_input = true
			continue
		}
		newline_input = false

		var found_correct_guess bool
		for i, r := range target {
			if guess == r && !correctly_guessed_runes[i] {
				found_correct_guess = true
				correctly_guessed_runes[i] = true
				number_of_correct_guesses++
			}
		}
		if found_correct_guess {
			fmt.Println("Correct!")
		} else {
			fmt.Println("Wrong!")
			remaining_tries--
		}

	}

	if remaining_tries > 0 {
		fmt.Println("Player 2 won. The word was:", target)
	} else {
		fmt.Println("Player 1 won. The word was:", target)
	}

}


