package logger

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	file, err := os.OpenFile("taskmaster.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(file, "TASKMASTER: ", log.LstdFlags)
}

func Log(msg string) {
	logger.Println(msg)
}