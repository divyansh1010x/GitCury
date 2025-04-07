package utils

import (
	"fmt"
	"log"
	"runtime"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

var LogLevel string = "info"

func SetLogLevel(level string) {
	LogLevel = level
}

// Debug logs a message at the debug level
func Debug(message string) {
	if LogLevel != "debug" {
		return
	}
	log.Printf("\n%s[DEBUG] 🐛 %s %s\n", Blue, Reset, message)
}

// Info logs a message at the info level
func Info(message string) {
	fmt.Printf("\n%s[INFO] ℹ️ %s %s\n", Green, Reset, message)
}

// Error logs error message
func Error(message string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		log.Printf("\n%s[ERROR] ❌%s %s (at %s:%d)\n", Red, Reset, message, file, line)
	} else {
		log.Printf("\n%s[ERROR] ❌%s %s\n", Red, Reset, message)
	}
}

// Print outputs data to the CLI
func Print(data string) {
	fmt.Printf("\n%s\n", data)
}
