package output

import (
	"GitCury/config"
	"GitCury/utils"
	"encoding/json"
	"os"
	"sync"
)

type FileEntry struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type Folder struct {
	Name  string      `json:"name"`
	Files []FileEntry `json:"files"`
}

type OutputData struct {
	Folders []Folder `json:"folders"`
}

var (
	outputData = OutputData{}
	mu         sync.RWMutex
)

func init() {
	LoadOutput()
}

func LoadOutput() {
	outputFilePath, ok := config.Get("output_file_path").(string)
	if !ok || outputFilePath == "" {
		outputFilePath = os.Getenv("HOME") + "/.gitcury/output.json"
	}

	file, err := os.Open(outputFilePath)
	if os.IsNotExist(err) {
		utils.Debug("[" + config.Aliases.Output + "]: No existing output file found. Initializing fresh output.")
		return
	} else if err != nil {
		utils.Error("[" + config.Aliases.Output + "]: 🚨 Error loading output file: " + err.Error())
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&outputData); err != nil {
		utils.Error("[" + config.Aliases.Output + "]: 🚨 Error decoding output file: " + err.Error())
	}

	utils.Debug("[" + config.Aliases.Output + "]: Loaded output data successfully.")
}

func Set(file, rootFolder string, commitMessage string) {
	mu.Lock()
	defer mu.Unlock()

	utils.Debug("[" + config.Aliases.Output + "]: Setting commit message for file: " + file + " in folder: " + rootFolder)
	folder := findOrCreateFolder(rootFolder)

	updated := false
	for i, entry := range folder.Files {
		if entry.Name == file {
			folder.Files[i].Message = commitMessage
			updated = true
			break
		}
	}

	if !updated {
		folder.Files = append(folder.Files, FileEntry{Name: file, Message: commitMessage})
	}

	utils.Debug("[" + config.Aliases.Output + "]: Commit message set for file: " + file + " in folder: " + rootFolder)
}

func Get(file string, rootFolder string) string {
	mu.RLock()
	defer mu.RUnlock()

	folder := findFolder(rootFolder)
	if folder == nil {
		return ""
	}

	for _, entry := range folder.Files {
		if entry.Name == file {
			return entry.Message
		}
	}
	return ""
}

func GetFolder(rootFolder string) Folder {
	mu.RLock()
	defer mu.RUnlock()

	folder := findFolder(rootFolder)
	if folder != nil {
		return *folder
	}
	return Folder{Name: rootFolder, Files: []FileEntry{}}
}

func GetAll() OutputData {
	mu.RLock()
	defer mu.RUnlock()

	copy := OutputData{Folders: make([]Folder, len(outputData.Folders))}
	for i, folder := range outputData.Folders {
		copy.Folders[i] = Folder{
			Name:  folder.Name,
			Files: append([]FileEntry{}, folder.Files...),
		}
	}
	return copy
}

func Delete(file string, rootFolder string) {
	mu.Lock()
	defer mu.Unlock()

	folder := findFolder(rootFolder)
	if folder == nil {
		utils.Error("[" + config.Aliases.Output + "]: ⚠️ Folder not found: " + rootFolder)
		return
	}

	for i, entry := range folder.Files {
		if entry.Name == file {
			folder.Files = append(folder.Files[:i], folder.Files[i+1:]...)
			break
		}
	}

	if len(folder.Files) == 0 {
		RemoveFolder(rootFolder)
	}

	SaveToFile()
	utils.Debug("[" + config.Aliases.Output + "]: File deleted and output saved.")
}

func Clear() {
	mu.Lock()
	defer mu.Unlock()
	outputData = OutputData{}

	outputFilePath, ok := config.Get("output_file_path").(string)
	if !ok || outputFilePath == "" {
		outputFilePath = os.Getenv("HOME") + "/.gitcury/output.json"
	}

	if err := os.Remove(outputFilePath); err != nil && !os.IsNotExist(err) {
		utils.Error("[" + config.Aliases.Output + "]: 🚨 Error deleting output file: " + err.Error())
	} else {
		utils.Debug("[" + config.Aliases.Output + "]: Output file cleared successfully.")
	}
}

func SaveToFile() {
	utils.Debug("[" + config.Aliases.Output + "]: Saving output data to file...")
	mu.RLock()
	defer mu.RUnlock()

	outputFilePath, ok := config.Get("output_file_path").(string)
	if !ok || outputFilePath == "" {
		outputFilePath = os.Getenv("HOME") + "/.gitcury/output.json"
		config.Set("output_file_path", outputFilePath)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utils.Error("[" + config.Aliases.Output + "]: 🚨 Error saving output file: " + err.Error())
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(outputData); err != nil {
		utils.Error("[" + config.Aliases.Output + "]: 🚨 Error encoding output data: " + err.Error())
	}

	utils.Debug("[" + config.Aliases.Output + "]: Output data saved successfully to: " + outputFilePath)
}

func findFolder(name string) *Folder {
	for i := range outputData.Folders {
		if outputData.Folders[i].Name == name {
			return &outputData.Folders[i]
		}
	}
	return nil
}

func findOrCreateFolder(name string) *Folder {
	folder := findFolder(name)
	if folder == nil {
		outputData.Folders = append(outputData.Folders, Folder{Name: name, Files: []FileEntry{}})
		return &outputData.Folders[len(outputData.Folders)-1]
	}
	return folder
}

func RemoveFolder(name string) {
	for i, folder := range outputData.Folders {
		if folder.Name == name {
			outputData.Folders = append(outputData.Folders[:i], outputData.Folders[i+1:]...)
			break
		}
	}

	SaveToFile()
	utils.Debug("[" + config.Aliases.Output + "]: Folder removed and output saved.")
}
