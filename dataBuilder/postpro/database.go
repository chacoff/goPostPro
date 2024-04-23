package postpro

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database_Line struct {
	timestamp string
	max_Tr1   int64
	mean_Tr1  int64
	mean_Web  int64
	min_Web   int64
	max_Tr3   int64
	mean_Tr3  int64
	width     int64
	threshold int64
	filename  string
}

func (database_line *Database_Line) Import_line_processing(line_processing LineProcessing) {
	database_line.timestamp = line_processing.timestamp.Format("15:04:05.000 ")
	database_line.max_Tr1 = int64(line_processing.max_Tr1)
	database_line.mean_Tr1 = int64(line_processing.mean_Tr1)
	database_line.mean_Web = int64(line_processing.mean_Web)
	database_line.min_Web = int64(line_processing.min_Web)
	database_line.max_Tr3 = int64(line_processing.max_Tr3)
	database_line.mean_Tr3 = int64(line_processing.mean_Tr3)
	database_line.width = int64(line_processing.width)
	database_line.threshold = int64(line_processing.threshold)
	database_line.filename = line_processing.filename

}

type CalculationsDatabase struct {
	database *sql.DB
}

func (calculations_database *CalculationsDatabase) Open_database() error {
	database, opening_error := sql.Open("sqlite3", DATABASE_PATH)
	if opening_error != nil {
		log.Println(opening_error)
		return opening_error
	}
	calculations_database.database = database
	return nil
}

func (calculations_database *CalculationsDatabase) Insert_line(line Database_Line) error {
	preparation, preparation_error := calculations_database.database.Prepare(
		"INSERT INTO Measures(Timestamp, Tr1_Max, Tr1_Mean, Web_Mean, Web_Min, Tr3_Max, Tr3_Mean, Width, Threshold, Filename) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if preparation_error != nil {
		return preparation_error
	}
	defer preparation.Close()
	// Execute it with the given values
	_, execution_error := preparation.Exec(line.timestamp, line.max_Tr1, line.mean_Tr1, line.mean_Web, line.min_Web, line.max_Tr3, line.mean_Tr3, line.width, line.threshold, line.filename)
	if execution_error != nil {
		return execution_error
	}
	return nil
}
