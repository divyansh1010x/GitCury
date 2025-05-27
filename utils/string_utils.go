package utils

import (
"strings"
)

// ContainsString checks if a string contains a substring
func ContainsString(s, substr string) bool {
return strings.Contains(s, substr)
}
