package watcher

import (
	"bufio"
	"log"
	"goPostPro/postpro"
	"os"
)

// FileReader reads files and return line by line
func FileReader(filePath string, fileName string) int {

	file, err := os.Open(filePath)
	if err != nil {
		log.Println("[WATCHER] errors opening the file:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip first line
	lineCount := 0

	// Read and process each line
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		processError := postpro.Process_line(line, fileName)
		if processError != nil {
			log.Println("[ERROR]", processError)
			return lineCount
		}
	}
	log.Println("[WATCHER] processed: ", fileName)

	// Scanner error
	if err := scanner.Err(); err != nil {
		log.Println("[WATCHER] errors reading the file: ", err)
	}
	return lineCount
}
