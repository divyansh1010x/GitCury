package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BinaryFileExtensions contains common binary file extensions
var BinaryFileExtensions = map[string]bool{
	// Images
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".tiff": true, ".svg": true, ".webp": true, ".ico": true,
	// Videos
	".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true, ".mkv": true, ".m4v": true,
	// Audio
	".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".wma": true, ".m4a": true,
	// Archives
	".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
	// Executables
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".app": true, ".deb": true, ".rpm": true,
	// Documents (binary formats)
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	// Fonts
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	// Database files
	".db": true, ".sqlite": true, ".sqlite3": true,
	// Other binary formats
	".bin": true, ".dat": true, ".dump": true, ".img": true, ".iso": true, ".dmg": true,
}

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

// IsBinaryFile checks if a file is likely binary based on its extension
func IsBinaryFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return BinaryFileExtensions[ext]
}

// GetBinaryFileType returns a descriptive type for binary files
func GetBinaryFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	// Images
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg", ".webp", ".ico":
		return "image"
	// Videos
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v":
		return "video"
	// Audio
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a":
		return "audio"
	// Archives
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz":
		return "archive"
	// Executables
	case ".exe", ".dll", ".so", ".dylib", ".app", ".deb", ".rpm":
		return "executable"
	// Documents
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return "document"
	// Fonts
	case ".ttf", ".otf", ".woff", ".woff2", ".eot":
		return "font"
	// Database
	case ".db", ".sqlite", ".sqlite3":
		return "database"
	default:
		return "binary"
	}
}

// GenerateBinaryCommitMessage creates appropriate commit messages for binary files
func GenerateBinaryCommitMessage(filePath string, gitStatus string) string {
	fileName := filepath.Base(filePath)
	fileType := GetBinaryFileType(filePath)

	switch strings.TrimSpace(gitStatus) {
	case "A", "??": // Added/new file
		switch fileType {
		case "image":
			return fmt.Sprintf("Add image asset: %s", fileName)
		case "video":
			return fmt.Sprintf("Add video asset: %s", fileName)
		case "audio":
			return fmt.Sprintf("Add audio asset: %s", fileName)
		case "archive":
			return fmt.Sprintf("Add archive: %s", fileName)
		case "executable":
			return fmt.Sprintf("Add executable: %s", fileName)
		case "document":
			return fmt.Sprintf("Add document: %s", fileName)
		case "font":
			return fmt.Sprintf("Add font asset: %s", fileName)
		case "database":
			return fmt.Sprintf("Add database file: %s", fileName)
		default:
			return fmt.Sprintf("Add binary file: %s", fileName)
		}
	case "M": // Modified
		switch fileType {
		case "image":
			return fmt.Sprintf("Update image asset: %s", fileName)
		case "video":
			return fmt.Sprintf("Update video asset: %s", fileName)
		case "audio":
			return fmt.Sprintf("Update audio asset: %s", fileName)
		case "archive":
			return fmt.Sprintf("Update archive: %s", fileName)
		case "executable":
			return fmt.Sprintf("Update executable: %s", fileName)
		case "document":
			return fmt.Sprintf("Update document: %s", fileName)
		case "font":
			return fmt.Sprintf("Update font asset: %s", fileName)
		case "database":
			return fmt.Sprintf("Update database file: %s", fileName)
		default:
			return fmt.Sprintf("Update binary file: %s", fileName)
		}
	case "D": // Deleted
		switch fileType {
		case "image":
			return fmt.Sprintf("Remove image asset: %s", fileName)
		case "video":
			return fmt.Sprintf("Remove video asset: %s", fileName)
		case "audio":
			return fmt.Sprintf("Remove audio asset: %s", fileName)
		case "archive":
			return fmt.Sprintf("Remove archive: %s", fileName)
		case "executable":
			return fmt.Sprintf("Remove executable: %s", fileName)
		case "document":
			return fmt.Sprintf("Remove document: %s", fileName)
		case "font":
			return fmt.Sprintf("Remove font asset: %s", fileName)
		case "database":
			return fmt.Sprintf("Remove database file: %s", fileName)
		default:
			return fmt.Sprintf("Remove binary file: %s", fileName)
		}
	default:
		return fmt.Sprintf("Update %s: %s", fileType, fileName)
	}
}
