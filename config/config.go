package config

import (
	"GitCury/utils"
	"encoding/json"
	"os"
	"sync"
)

var (
	settings = make(map[string]interface{})
	mu       sync.RWMutex
)

func init() {
	LoadConfig()
}

func LoadConfig() {
	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		// Set default settings if the file does not exist
		settings = map[string]interface{}{
			"app_name": "GitCury",
			"version":  "1.0.0",
			"root_folders": []string{"."},
		}
		utils.Error("Config file not found. Using default settings: \n" + utils.ToJSON(settings))
		return
	} else if err != nil {
		utils.Error("Error loading config file: " + err.Error())
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		utils.Error("Error decoding config file: " + err.Error())
		panic(err)
	}

	utils.Debug("Loaded Settings:\n" + utils.ToJSON(settings))
}

func Set(key string, value interface{}) {
	mu.Lock()
	defer mu.Unlock()
	settings[key] = value

	go func() {
		mu.RLock()
		defer mu.RUnlock()

		file, err := os.OpenFile("config.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			utils.Error("Error saving config: " + err.Error())
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(settings); err != nil {
			utils.Error("Error saving config: " + err.Error())
		}
	}()
}

func Get(key string) interface{} {
	mu.RLock()
	defer mu.RUnlock()
	return settings[key]
}

func GetAll() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()
	copy := make(map[string]interface{})
	for key, value := range settings {
		copy[key] = value
	}
	return copy
}
