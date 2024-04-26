package watcher

import (
	"bufio"
	"fmt"
	"goPostPro/postpro"
	"os"
)

//FileReader reads files and return line by line
func FileReader(file_path string, filename string) int {

	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	lineCount := 0

	// Read and process each line
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		process_error := postpro.Process_line(line, filename)
		if process_error != nil {
			fmt.Println(process_error)
			return lineCount
		}
	}

	// Scanner error
	if err := scanner.Err(); err != nil {
		fmt.Println("Erreur lors de la lecture du fichier:", err)
	}
	return lineCount
}


