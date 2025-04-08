package utils

import (
	"encoding/json"
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
