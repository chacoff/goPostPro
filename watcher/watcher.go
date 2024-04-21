package watcher

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"goPostPro/global"
	"os"
)

func Watcher() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event.Op, event.Name)
				if event.Has(fsnotify.Write) {
					fmt.Println("modified file:", event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	folders, errR := readFoldersFromJSON(global.Appconfig.DataFolders)
	if errR != nil {
		fmt.Println("Error reading folders from JSON file:", errR)
	}

	// Add a paths
	for _, folder := range folders {

		_, errF := os.Stat(folder)
		if os.IsNotExist(errF) {
			errM := os.MkdirAll(folder, 0755)
			if errM != nil {
				fmt.Println("Error creating folder: ", errM)
			}
		}

		err = watcher.Add(folder)
		if err != nil {
			fmt.Println("watcher.Add error: ", err)
		}
	}
	fmt.Println("Observing folders")
	// Block main goroutine forever.
	<-make(chan struct{})
}

func readFoldersFromJSON(filePath string) ([]string, error) {
	var folders []string

	// Read JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON data
	err = json.Unmarshal(data, &folders)
	if err != nil {
		return nil, err
	}

	return folders, nil
}
