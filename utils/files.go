package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func ListFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		Debug("[FILES]: 🚨 Error reading directory: " + err.Error())
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	Debug("[FILES]: 📂 Successfully listed files in directory: " + directory)
	return files, nil
}

func ToJSON(data interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		Debug("[JSON]: 🚨 Error marshalling data: " + err.Error())
		return "{}"
	}
	Debug("[JSON]: ✨ Successfully marshalled data to JSON")
	return string(jsonData)
}

func IsNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}

	Debug("[NUMERIC]: 🔢 String is numeric: " + s)
	return len(s) > 0
}

func ParseInt(s string) (int, error) {
	if !IsNumeric(s) {
		Error("[PARSE]: 🚨 Error parsing string to int: " + s)
		return 0, fmt.Errorf("invalid number: %s", s)
	}

	result := 0
	for _, char := range s {
		result = result*10 + int(char-'0')
	}

	Debug("[PARSE]: 🔢 Successfully parsed string to int: " + s)
	return result, nil
}

func ParseFloat(s string) (float64, error) {
	var result float64
	if _, err := fmt.Sscanf(s, "%f", &result); err != nil {
		Error("[PARSE]: 🚨 Error parsing string to float: " + s)
		return 0, fmt.Errorf("invalid float: %s", s)
	}

	Debug("[PARSE]: 🔢 Successfully parsed string to float: " + s)
	return result, nil
}
