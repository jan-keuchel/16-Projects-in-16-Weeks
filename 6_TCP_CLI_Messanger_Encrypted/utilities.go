package main

import (
	"fmt"
	"regexp"
	"unicode"
)

func isNumeric(s string) bool {

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true

}


func verifyIPFormat(ip string) bool {

	ipPattern := `(^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$)|localhost`

	re, err := regexp.Compile(ipPattern)
	if err != nil {
		fmt.Println("[Error] Compiling RegEx failed:", err)
		return false
	}

	return re.MatchString(ip)

}
