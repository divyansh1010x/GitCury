package output

import (
	"GitCury/utils"
	"encoding/json"
	"os"
	"sync"
)

var (
	files = make(map[string]string) // Map to store file as key and commit message as value
	mu    sync.RWMutex
)

func init() {
	LoadOutput()
}

func LoadOutput() {
	file, err := os.Open("output.json")
	if os.IsNotExist(err) {
		utils.Debug("Output file not found")
		return
	} else if err != nil {
		utils.Error("Error loading output file: " + err.Error())
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&files); err != nil {
		utils.Error("Error decoding output file: " + err.Error())
	}

	convertedFiles := make(map[string]interface{})
	for key, value := range files {
		convertedFiles[key] = value
	}
	utils.Debug("Loaded output:\n" + utils.ToJSON(convertedFiles))
}

func Set(file, commitMessage string) {
	mu.Lock()
	defer mu.Unlock()
	files[file] = commitMessage

	go func() {
		mu.RLock()
		defer mu.RUnlock()

		outputFile, err := os.OpenFile("output.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			utils.Error("Error saving output file: " + err.Error())
			return
		}
		defer outputFile.Close()

		encoder := json.NewEncoder(outputFile)
		if err := encoder.Encode(files); err != nil {
			utils.Error("Error saving output file: " + err.Error())
		}
	}()
}

func Get(file string) string {
	mu.RLock()
	defer mu.RUnlock()
	return files[file]
}

func GetAll() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	copy := make(map[string]string)
	for key, value := range files {
		copy[key] = value
	}
	return copy
}

func Delete(file string) {
	mu.Lock()
	defer mu.Unlock()
	delete(files, file)

	go func() {
		mu.RLock()
		defer mu.RUnlock()

		outputFile, err := os.OpenFile("output.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			utils.Error("Error saving output file: " + err.Error())
			return
		}
		defer outputFile.Close()

		encoder := json.NewEncoder(outputFile)
		if err := encoder.Encode(files); err != nil {
			utils.Error("Error saving output file: " + err.Error())
		}
	}()
}

func Clear() {
	mu.Lock()
	defer mu.Unlock()
	files = make(map[string]string)

	if err := os.Remove("output.json"); err != nil && !os.IsNotExist(err) {
		utils.Error("Error deleting output file: " + err.Error())
	}
}
