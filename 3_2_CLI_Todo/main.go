package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"strconv"
)

// Entry is a structure that holds a todo-item in a todo-list.
// Every entry has a name, a date on which it has to be done, a priority: 1, 2, 3 (1 being highest)
// and a status: 0, 1, 2 (0 := todo, 1 := working on it, 2 := done)
type Entry struct {
	name 		 string
	due_date time.Time
	priority uint8
	status 	 uint8
}

func (e Entry) write_entry(list_name string) int {

	var builder strings.Builder
	builder.WriteString(e.name)
	builder.WriteString(",")
	
	builder.WriteString(e.due_date.Format("02.01.2006"))
	builder.WriteString(",")

	builder.WriteString(strconv.FormatUint(uint64(e.priority), 10))
	builder.WriteString(",")

	builder.WriteString(strconv.FormatUint(uint64(e.status), 10))
	builder.WriteString("\n")

	file, err := os.OpenFile(list_name, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(builder.String())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully added todo '%s' to '%s'\n", e.name, list_name)
	return 0

}

func main() {
	
	if len(os.Args) == 1 {
		fmt.Printf("Use the programm as following:\n./codo -<arg1> [val1] -<arg2> [val2] [...]\n\n")
		fmt.Printf("The following flags and arguments exist:\n")
		fmt.Printf("-c <list_name>: Create new todo list\n")
		fmt.Printf("-d <list_name>: Delete todo list\n")
		fmt.Printf("-l <list_name> -p number: List todo items of the list with a given priority. 0 for every item.\n")
	}

	// for i, arg := range os.Args {
	// 	fmt.Printf("Arg-%v: %v\n", i, arg)
	// }

	n := len(os.Args)
	for i, arg := range os.Args {

		if arg[0] == '-' {
			switch arg[1] {
			case 'c':
				if i + 1 < n {
					fmt.Printf("Create file...\n")
					create_list(os.Args[i+1])
				}
			case 'd':
				if i + 1 < n {
					fmt.Printf("Delete file...\n")
					delete_list(os.Args[i+1])
				}
			case 'l':
				if i + 2 < n {
					if os.Args[i+2][0] == '-' && os.Args[i+2][1] == 'p' {
						fmt.Printf("List list with priority %v\n", os.Args[i+2][1])
					} else {
						fmt.Printf("Please provide a priority to filter by.\n")
					}
				}
			case 'a':
				if i + 5 < n {
					t, err := time.Parse("02.01.2006", os.Args[i+3])
					if err != nil {
						fmt.Printf("Error while parsing string to date.\n")
						log.Fatal(err)
					}
					priority, err := strconv.ParseUint(os.Args[i+4], 10, 8)
					if err != nil {
						fmt.Println("Error parsing string to uint8 (priority):\n", err)
						return
					}
					status, err := strconv.ParseUint(os.Args[i+5], 10, 8)
					if err != nil {
						fmt.Println("Error parsing string to uint8 (status):\n", err)
						return
					}
					e := Entry{ os.Args[i+2], t, uint8(priority), uint8(status) }
					e.write_entry(os.Args[i+1])
				}
			}
		}

	}




}

// create_list creates a file of name 'list_name' in which the todos are stored.
// Returns 0 on success, 1 if the list already exists, 2 otherwise.
func create_list(list_name string) int {

	if list_exists(list_name) {
		fmt.Printf("A list with the name '%s' already exists. Couldn't create another one.\n", list_name)
		return 1
	}

	file, err := os.Create(list_name)
	if err != nil {
		fmt.Printf("Error while creating a list:\n")
		log.Fatal(err)
	}
	defer file.Close()

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
