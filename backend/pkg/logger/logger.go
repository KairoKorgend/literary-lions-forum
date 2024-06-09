package logger

import (
	"io"
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	infoPrefix := "\033[1;34mINFO:\033[0m "   // Blue color
	errorPrefix := "\033[1;31mERROR:\033[0m " // Red color

	// Check if the IS_TEST_ENVIRONMENT environment variable is set
	if os.Getenv("IS_TEST_ENVIRONMENT") == "true" {
		// If it is, discard all log output
		InfoLogger = log.New(io.Discard, infoPrefix, log.Ldate|log.Ltime|log.Lshortfile)
		ErrorLogger = log.New(io.Discard, errorPrefix, log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		InfoLogger = log.New(os.Stdout, infoPrefix, log.Ldate|log.Ltime|log.Lshortfile)
		ErrorLogger = log.New(os.Stderr, errorPrefix, log.Ldate|log.Ltime|log.Lshortfile)
	}
}
