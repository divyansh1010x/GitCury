package utils

import (
	"encoding/json"
	"os"
)

func ListFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		Error("Error reading directory: " + err.Error())
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func ToJSON(data map[string]interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		Error("Error marshalling data: " + err.Error())
		return "{}"
	}
	return string(jsonData)
}
