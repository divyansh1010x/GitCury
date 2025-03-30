package output

import (
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
	if err := decoder.Decode(&outputData); err != nil {
		utils.Error("Error decoding output file: " + err.Error())
	}

	utils.Debug("Loaded output:\n" + utils.ToJSON(outputData))
}

func Set(file, rootFolder string, commitMessage string) {
	mu.Lock()
	defer mu.Unlock()

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

	saveToFile()
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

func GetAllFiles(rootFolder string) Folder {
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
		removeFolder(rootFolder)
	}

	saveToFile()
}

func Clear() {
	mu.Lock()
	defer mu.Unlock()
	outputData = OutputData{}

	if err := os.Remove("output.json"); err != nil && !os.IsNotExist(err) {
		utils.Error("Error deleting output file: " + err.Error())
	}
}

func saveToFile() {
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
		if err := encoder.Encode(outputData); err != nil {
			utils.Error("Error saving output file: " + err.Error())
		}
	}()
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

func removeFolder(name string) {
	for i, folder := range outputData.Folders {
		if folder.Name == name {
			outputData.Folders = append(outputData.Folders[:i], outputData.Folders[i+1:]...)
			break
		}
	}
}
