package main

import (
	"fmt"
	"os"
	"log"
)

func main() {
	fmt.Printf("Test of file creation and manipulation:\n")

	file, err := os.Create("Test")
	if err != nil {
		log.Fatal(err) // prints the error message and exits with error code.
	}
	defer file.Close() // the statement behind 'defer' will be executed at the end of the function from within the defer statement is made. Multiple 'defer' statements will be called according to LIFO (Stack).

	data := []byte("This is a test.\n")
	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	// Qick variant:
	// err := os.WriteFile("example.txt", []byte("Quick write!\n"), 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	writing()

	read := reading()
	fmt.Printf("Read from file: \n%v", read)
	

}

func writing() {

	file, err := os.OpenFile("Test", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString("Appending this line.\n")
	if err != nil {
		log.Fatal(err)
	}

}

func reading() string {

	bytes, err := os.ReadFile("Test")
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)

}
