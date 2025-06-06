package main

import (
	"fmt"
	"os"
	"regexp"
	"unicode"
)

type Packet struct {
	MsgType string `json:"msgType"`
	Payload	string `json:"payload"`
}

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

func fileExists(path string) bool {

    _, err := os.Stat(path)
    if err == nil {
        return true
    }
    if os.IsNotExist(err) {
        return false
    }
    return false

}

func getFileSize(path string) (int, error) {

	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(fileInfo.Size()), nil

}
