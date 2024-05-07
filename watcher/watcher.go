package watcher

import (
	"encoding/json"
	"log"
	"goPostPro/global"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func Watcher() {
	watcher, err := fsnotify.NewWatcher() // create new watcher
	if err != nil {
		log.Println(err)
	}
	defer watcher.Close()

	// start listening for events
	go func() {
		var (
			timer     *time.Timer
			lastEvent fsnotify.Event
		)

		timer = time.NewTimer(time.Millisecond)
		<-timer.C // timer should be expired at first

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				lastEvent = event
				timer.Reset(time.Duration(global.Appconfig.TickWatcher) * time.Millisecond)
			case errW, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("[WATCHER] error:", errW)
			case <-timer.C:
				if lastEvent.Op&fsnotify.Write == fsnotify.Write {
					onModified(lastEvent.Name)
				} else if lastEvent.Op&fsnotify.Create == fsnotify.Create {
					onCreate(lastEvent.Name)
				} else if lastEvent.Op&fsnotify.Remove == fsnotify.Remove {
					onDelete(lastEvent.Name)
				}
				//if err != nil{
				//	log.Println("[WATCHER] error: ", err)
				//}
			}
		}
	}()

	// add/create folders to watch. we don't have a proper recursive function, yet we know the folder names in advance
	folders, errR := readFoldersFromJSON(global.Appconfig.DataFolders)
	if errR != nil {
		log.Println("Error reading folders from JSON file:", errR)
	}

	// add a paths
	for _, folder := range folders {
		_, errF := os.Stat(folder)
		if os.IsNotExist(errF) {
			errM := os.MkdirAll(folder, 0755)
			if errM != nil {
				log.Println("Error creating folder: ", errM)
			}
		}

		errWa := watcher.Add(folder)
		if errWa != nil {
			log.Println("watcher.Add error: ", errWa)
		}
	}
	log.Printf("[WATCHER] Observing folders in: %s\n", global.Appconfig.DataFolders)

	<-make(chan struct{}) // block main goroutine forever
}

// onModified handles the file reading
func onModified(fileAddress string) {
	fileNameList := fileAddress
	fileName := fileNameList[strings.LastIndex(fileNameList, "\\")+1:]
	log.Println("[WATCHER] modified file:", fileAddress)
	FileReader(fileAddress, fileName)
}

func onCreate(fileAddress string) {
	log.Println("[WATCHER] created file:", fileAddress)
}

func onDelete(fileAddress string) {
	log.Println("[WATCHER] deleted file:", fileAddress)
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
