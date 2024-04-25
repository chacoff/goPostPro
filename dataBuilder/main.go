package main

import (
	"bufio"
	"go_train2/postpro"
	"log"
	"os"
	"time"
)

func main() {
	calculations_database := postpro.CalculationsDatabase{}
	calculations_database.Open_database()
	calculations_database.Drop_Table()
	calculations_database.Create_Table()
	folder := "examples/"
	files := []string{"DUO01-02_0891.txt", "DUO01-02_0892.txt", "DUO01-02_0894.txt", "DUO01-02_0895.txt", "DUO01-02_0896.txt", "DUO01-02_0897.txt", "DUO01-02_0898.txt", "DUO01-02_0899.txt", "DUO01-02_0900.txt"}
	debut := time.Now()
	line_count := 0

	for _, filename := range files {
		line_count += test(folder+filename, filename)
	}
	calculations_database.Query_database("2024-02-13 11:06:29", "2024-02-13 11:07:40")

	log.Println("Total execution time : ", time.Since(debut))
	log.Println("Number of line processed :", line_count)
	log.Println("Mean execution time : ", time.Since(debut).Milliseconds()/int64(line_count), " ms")
}

func test(file_path string, filename string) int {

	file, err := os.Open(file_path)
	if err != nil {
		log.Println("Erreur lors de l'ouverture du fichier:", err)
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
		process_error := postpro.Process_line(line, filename)
		if process_error != nil {
			log.Println(process_error)
			return lineCount
		}
	}

	// Scanner error
	if err := scanner.Err(); err != nil {
		log.Println("Erreur lors de la lecture du fichier:", err)
	}
	return lineCount
}
