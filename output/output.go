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
		// utils.Error("Output file not found")
		return
	} else if err != nil {
		utils.Error("Error loading output file: " + err.Error())
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&outputData); err != nil {
		utils.Error("Error decoding output file: " + err.Error())
	}

	// utils.Debug("Loaded output:\n" + utils.ToJSON(outputData))
}

func Set(file, rootFolder string, commitMessage string) {
	mu.Lock()
	defer mu.Unlock()

	utils.Debug("Setting commit message for file: " + file + " in folder: " + rootFolder)
	// rootFolder := getRootFolder(file)
	folder := findOrCreateFolder(rootFolder)

	// Check if the file already exists in the folder and update it
	updated := false
	for i, entry := range folder.Files {
		if entry.Name == file {
			folder.Files[i].Message = commitMessage
			updated = true
			break
		}
	}

	// If not updated, add a new file entry
	if !updated {
		folder.Files = append(folder.Files, FileEntry{Name: file, Message: commitMessage})
	}

	utils.Debug("Set commit message for file: " + file + " in folder: " + rootFolder)
}

func Get(file string, rootFolder string) string {
	mu.RLock()
	defer mu.RUnlock()

	// rootFolder := getRootFolder(file)
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

	// rootFolder := getRootFolder(file)
	folder := findFolder(rootFolder)
	if folder == nil {
		utils.Error("Folder not found: " + rootFolder)
		return
	}

	for i, entry := range folder.Files {
		if entry.Name == file {
			folder.Files = append(folder.Files[:i], folder.Files[i+1:]...)
			break
		}
	}

	// Remove the folder if it's empty
	if len(folder.Files) == 0 {
		RemoveFolder(rootFolder)
	}

	SaveToFile()
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
		utils.Error("Error deleting output file: " + err.Error())
	}
}

func SaveToFile() {
	utils.Debug("Saving output to file... with data: " + utils.ToJSON(outputData))
	mu.RLock()
	defer mu.RUnlock()

	outputFilePath, ok := config.Get("output_file_path").(string)
	if !ok || outputFilePath == "" {
		outputFilePath = os.Getenv("HOME") + "/.gitcury/output.json"
		config.Set("output_file_path", outputFilePath)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utils.Error("Error saving output file: " + err.Error())
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ") // Format JSON with indentation
	if err := encoder.Encode(outputData); err != nil {
		utils.Error("Error saving output file: " + err.Error())
	}

	outputDataJSON, _ := json.Marshal(outputData)
	utils.Debug("Output : " + string(outputDataJSON) + " saved to file: " + outputFilePath)
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
}
