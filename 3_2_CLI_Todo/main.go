package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	
	if len(os.Args) == 1 {
		fmt.Printf("Use the programm as following:\n./codo -<arg1> [val1] -<arg2> [val2] [...]\n\n")
		fmt.Printf("The following flags and arguments exist:\n")
		fmt.Printf("-c list_name: Create new todo list\n")
		fmt.Printf("-d list_name: Delete todo list\n")
		fmt.Printf("-l list_name -p number: List todo items of the list with a given priority. 0 for every item.\n")
	}

}

// create_list creates a file of name 'list_name' in which the todos are stored.
// Returns 0 on success, 1 if the list already exists, 2 otherwise.
func create_list(list_name string) int {

	if list_exists(list_name) {
		fmt.Printf("A list with the name '%s' already exists. Couldn't create another one.\n", list_name)
		return 1
	}

	_, err := os.Create(list_name)
	if err != nil {
		fmt.Printf("Error while creating a list:\n")
		log.Fatal(err)
	}

	fmt.Printf("The list '%s' was successfully created.\n", list_name)

	return 0

}

// delete_list deletes a list with the name 'list_name'.
// Returns 0 if the list was successfully deleted, 1 if such list didn't exist
func delete_list(list_name string) int {

	if !list_exists(list_name) {
		fmt.Printf("The list '%s' doesn't exist.\n", list_name)
		return 1
	}

	err := os.Remove(list_name)
	if err != nil {
		fmt.Printf("Error while deleting a list:\n")
		log.Fatal(err)
	}

	fmt.Printf("The list '%s' was successfully deleted.\n", list_name)

	return 0

}

// print_list prints the contents of the list 'list_name' to stdout.
// If priority is 0, every item will be printed. Othewise only items with the given priority
// will be printed to stdout.
// Returns 0 on success, 1 if the list does not exist, 2 if an error occurs.
func print_list(list_name string, priority uint8) int {

	if !list_exists(list_name) {
		fmt.Printf("The list '%s' does not exist. Create it using ./codo -c %s.", list_name, list_name)
		return 1
	}

	if priority == 0 {
		// TODO: print the entire list
		fmt.Println("Printing the entire list")
	} else {
		// TODO: print every item with priority of 'priority'
		fmt.Printf("Printing every item with priority %d", priority)
	}

	return 0

}

// list_exists returns true if a given list exists in this directory, false otherwise.
func list_exists(list_name string) bool {

	_, err := os.Stat(list_name)
	return !os.IsNotExist(err)

}
