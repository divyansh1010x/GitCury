package config

import (
	"GitCury/utils"
)

// ResetConfig clears all configuration settings - used for testing only
func ResetConfig() {
	mu.Lock()
	defer mu.Unlock()
	settings = make(map[string]interface{})
	utils.Debug("[Config]: 🔄 Configuration reset for testing")
}
