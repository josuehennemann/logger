package main

import (
	"fmt"
	"github.com/josuehennemann/logger"
	"os"
)

var logFile *logger.Logger

func main() {

	filename := "my_custom.log"
	var err error
	//create a log in development environment
	logFile, err = logger.New(filename, logger.INFO|logger.ERROR, true)
	if err != nil {
		fmt.Println("failed create file", err.Error())
		os.Exit(1)
	}

	defer logFile.Close()
	txt := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor"
	logFile.Printf(logger.ERROR, "Writing %s", txt)
	logFile.Printf(logger.WARN, "Warning file not found") //not write
	logFile.Printf(logger.INFO, "%s", txt)
	logFile.Printf(logger.DEBUG, "Something function") //not write

	callMe()
}

func callMe() {
	logFile.Println(logger.ERROR, "Printing stack trace")
}
