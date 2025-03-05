package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"strconv"
	"encoding/csv"
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

// write_entry writes the values of a todo item (entry) to the todo list 'list_name'.
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

// print prints the values of an Entry to stdout
func (e Entry) print() {

	var prio string
	switch e.priority {
	case 1:
		prio = "Hoch"
	case 2:
		prio = "Mittel"
	case 3:
		prio = "Niedrig"
	}

	var stat string
	switch e.status {
	case 0:
		stat = "Todo"
	case 1:
		stat = "In Progress"
	case 2:
		stat = "Done"
	}

	fmt.Printf("Name: %-20s due-date: %-15v priority: %-10s status: %-10s\n",
		e.name,
		e.due_date,
		prio,
		stat)

}

func main() {
	
	if len(os.Args) == 1 {
		fmt.Printf("Use the programm as following:\n./codo -<arg1> [val1] -<arg2> [val2] [...]\n\n")
		fmt.Printf("The following flags and arguments exist:\n")
		fmt.Printf("-c <list_name>: Create new todo list\n")
		fmt.Printf("-d <list_name>: Delete todo list\n")
		fmt.Printf("-l <list_name> -p number: List todo items of the list with a given priority. 0 for every item.\n")
		fmt.Printf("-a <list_name> <due_date DD.MM.YYYY> <priority [1, 2, 3]> <status [0, 1, 2]>: Add an entry with name, due-date, priority and status to a list.\n")
		fmt.Printf("-s <list_name> <task_name> <new_status>: Change the status of an item in a certain list.\n")
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
				if i + 3 < n {
					if os.Args[i+2][0] == '-' && os.Args[i+2][1] == 'p' {
						priority, err := strconv.ParseUint(os.Args[i+3], 10, 8)
						if err != nil {
							fmt.Println("Error parsing string to uint8 (priority):\n", err)
							return
						}
						print_list(os.Args[i+1], uint8(priority))
					} 
				} else {
					fmt.Printf("Please provide a priority to filter by.\n")
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
			case 's':
				if i + 3 < n {
					update_task_status(os.Args[i+1], os.Args[i+2], os.Args[i+3])
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

// csv_to_strings takes in a file name of a csv file and returns a slice of string slices with the contents
// of the csv file.
func csv_to_strings(csv_name string) [][]string {

	file, err := os.Open(csv_name)
	if err != nil {
		fmt.Printf("Error while opening file at csv conversion:\n")
		log.Fatal(err)
	}

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error while reading csv file:\n")
		log.Fatal(err)
	}

	return records

}

func get_list_entries(list_name string) []Entry {

	var entries_s [][]string = csv_to_strings(list_name)

	var entries []Entry
	for _, entry := range entries_s {
		var e Entry
		e.name = entry[0]

		t, err := time.Parse("02.01.2006", entry[1])
		if err != nil {
			fmt.Printf("Error while parsing string to date.\n")
			log.Fatal(err)
		}
		e.due_date = t

		priority, err := strconv.ParseUint(entry[2], 10, 8)
		if err != nil {
			fmt.Println("Error parsing string to uint8 (priority):\n", err)
			log.Fatal(err)
		}
		e.priority = uint8(priority)

		status, err := strconv.ParseUint(entry[3], 10, 8)
		if err != nil {
			fmt.Println("Error parsing string to uint8 (status):\n", err)
			log.Fatal(err)
		}
		e.status = uint8(status)

		entries = append(entries, e)

	}

	return entries

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

	if priority < 4 {

		var entries []Entry = get_list_entries(list_name)
		
		for _, entry := range entries {
			if priority == 0 || priority == entry.priority {
				entry.print()
			}
		}

	} else {

		fmt.Printf("Invalid priority to filter by given. Can be 0, 1, 2 or 3.\n")
		
	}

	return 0

}

// update_task_status changes to status of the task 'task_name' in the list 'list_name' to 'status'.
func update_task_status(list_name string, task_name string, status string) {

	if !list_exists(list_name) {
		fmt.Printf("The list '%s' does not exist. Create it using ./codo -c %s.", list_name, list_name)
		return
	}

	var entries_s [][]string = csv_to_strings(list_name)

	for i, entry := range entries_s {
		if task_name == entry[0] {
			entries_s[i][3] = status
		}
	}

	file, err := os.Create(list_name)
	if err != nil {
		fmt.Printf("Error while opening file: update_task_status\n")
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(entries_s)
	if err != nil {
		fmt.Printf("Error writing to file: update_task_status\n")
		log.Fatal(err)
	}
	writer.Flush()

	fmt.Printf("Updated status of task '%s' in list '%s' successfully.\n", task_name, list_name)

}

// list_exists returns true if a given list exists in this directory, false otherwise.
func list_exists(list_name string) bool {

	_, err := os.Stat(list_name)
	return !os.IsNotExist(err)

}
