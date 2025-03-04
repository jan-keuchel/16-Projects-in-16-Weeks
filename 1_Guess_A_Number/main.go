package main

import (
	"fmt"
	"math/rand"
)

const RANGE = 200
const TRIES = 10

func main() {


	fmt.Println("Welcome to 'Guess a Number'. You'll receive 'higher'- and 'lower'-hints. The number range is 0 to", RANGE, "(exclusive). You'll have", TRIES, "tries.")

	var target uint8 = uint8(rand.Intn(RANGE))
	var guess uint8;
	var remaining_tries = TRIES

	fmt.Println("The number was chosen. Take a guess:")
	fmt.Scan(&guess)
	remaining_tries--

	for guess != target && remaining_tries > 0 {
		fmt.Println("That was not correct.", remaining_tries, "remaining.")
		if target > guess {
			fmt.Println("The target is larger.")
		} else {
			fmt.Println("The target is smaller.")
		}
		fmt.Println("The number was chosen. Take a guess:")
		fmt.Scan(&guess)
		remaining_tries--
	}

	if remaining_tries > 0 {
		fmt.Println("Correct! The number was", target)
	} else {
		fmt.Println("You lost. The number was", target)
	}

}
