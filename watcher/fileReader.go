package watcher

import (
	"bufio"
	"fmt"
	"goPostPro/postpro"
	"os"
)

// FileReader reads files and return line by line
func FileReader(file_path string, filename string) int {

	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("[WATCHER] errors opening the file:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	lineCount := 0

	// Read and process each line
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		postpro.DATABASE.Open_database()
		processError := postpro.Process_line(line, filename)
		if processError != nil {
			fmt.Println(processError)
			return lineCount
		}
	}

	// Scanner error
	if err := scanner.Err(); err != nil {
		fmt.Println("[WATCHER] errors reading the file:", err)
	}
	return lineCount
}
