package utils

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

// ANSI color codes - Dark Tech palette
const (
	Reset     = "\033[0m"
	Red       = "\033[38;5;196m" // Bright red
	Green     = "\033[38;5;46m"  // Neon green
	Yellow    = "\033[38;5;226m" // Warning yellow
	Blue      = "\033[38;5;33m"  // Electric blue
	Magenta   = "\033[38;5;201m" // Cyber pink
	Cyan      = "\033[38;5;51m"  // Holographic cyan
	Black     = "\033[38;5;236m" // Dark background
	Bold      = "\033[1m"
	Underline = "\033[4m"
	Blink     = "\033[5m" // Use sparingly!
	Dim       = "\033[2m"
	BlackBg   = "\033[48;5;235m" // Dark background
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
	log.Printf("\n%s%s[SCAN    ] 🔍 %s %s\n", Cyan, BlackBg, Reset, message)
}

// Info logs a message at the info level
func Info(message string) {
	fmt.Printf("\n%s%s[SYS     ] ⚡ %s %s\n", Green, BlackBg, Reset, message)
}

// Success logs a success message
func Success(message string) {
	fmt.Printf("\n%s%s[SUCCESS ] 💻 %s %s\n", Green, BlackBg, Reset, message)
}

// Error logs error message
func Error(message string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		shortFile := file
		if parts := strings.Split(file, "/"); len(parts) > 2 {
			shortFile = parts[len(parts)-2] + "/" + parts[len(parts)-1]
		}
		log.Printf("\n%s%s[BREACH  ] ⚠️ %s %s (at %s:%d)\n", Red, BlackBg, Reset, message, shortFile, line)
	} else {
		log.Printf("\n%s%s[BREACH  ] ⚠️ %s %s\n", Red, BlackBg, Reset, message)
	}
}

// Warning logs warning message
func Warning(message string) {
	fmt.Printf("\n%s%s[ALERT   ] 🔥 %s %s\n", Yellow, BlackBg, Reset, message)
}

// Print outputs data to the CLI
func Print(data string) {
	fmt.Printf("\n%s%s%s\n", Cyan, data, Reset)
}
