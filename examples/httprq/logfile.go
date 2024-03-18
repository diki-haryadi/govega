package main

import (
	"fmt"
	"log"
	"os"
)

func LogToFile(logName string, data string) error {
	file, errFile := os.OpenFile("custom.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errFile != nil {
		fmt.Println(errFile.Error())
	}

	logger := log.New(file, logName+" > ", log.LstdFlags)
	logger.Println(data)
	logger.Println("-")
	file.Close()
	return nil
}
