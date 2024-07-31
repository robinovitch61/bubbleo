package dev

import (
	"log"
	"os"
)

func Debug(msg string) {
	debugPath := "viewport.log"
	file, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger := log.New(file, "", log.Ldate|log.Lmicroseconds)
	logger.Printf("%q", msg)
}
