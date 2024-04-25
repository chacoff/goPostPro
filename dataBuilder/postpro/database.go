package postpro

import (
	"database/sql"
	"log"
	"time"

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
	database_line.timestamp = line_processing.timestamp.Format(TIME_FORMAT)
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

func (calculations_database *CalculationsDatabase) Create_Table() error {
	_, query_error := calculations_database.database.Exec(`
		CREATE TABLE IF NOT EXISTS Measures (
			Timestamp TEXT,
			Tr1_Max   INTEGER,
			Tr1_Mean  INTEGER,
			Web_Mean  INTEGER,
			Web_Min   INTEGER,
			Tr3_Max   INTEGER,
			Tr3_Mean  INTEGER,
			Width     INTEGER,
			Threshold INTEGER,
			Filename  TEXT
		);`)
	return query_error
}

func (calculations_database *CalculationsDatabase) Drop_Table() error {
	_, query_error := calculations_database.database.Exec(`DROP TABLE IF EXISTS Measures;`)
	return query_error
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

func (calculations_database *CalculationsDatabase) Query_database(begin_string_timestamp string, end_string_timestamp string) error {
	begin_timestamp, parsing_error := time.Parse(TIME_FORMAT_REQUESTS, begin_string_timestamp)
	if parsing_error != nil {
		return parsing_error
	}
	end_timestamp, parsing_error := time.Parse(TIME_FORMAT_REQUESTS, end_string_timestamp)
	if parsing_error != nil {
		return parsing_error
	}

	rows, err := calculations_database.database.Query(`
	SELECT
		MAX(Tr1_Max) AS Query_Tr1_Max,
		AVG(Tr1_Mean) AS Query_Tr1_Mean,
		Query_Web_Mean,
		MIN(Web_Min) AS Query_Web_Min,
		MAX(Tr3_Max) AS Query_Tr3_Max,
		AVG(Tr3_Mean) AS Query_Tr3_Mean,
		AVG((Web_Mean-Query_Web_Mean)*(Web_Mean-Query_Web_Mean)) AS Query_Web_Variance,
		AVG(Width) AS Query_Width_Mean,
		AVG(Threshold) AS Query_Threshold_Mean
	FROM Measures,
		(SELECT AVG(Web_Mean) AS Query_Web_Mean
		FROM Measures
		WHERE Timestamp BETWEEN '` + begin_timestamp.Format(TIME_FORMAT) + `' AND '` + end_timestamp.Format(TIME_FORMAT) + `')
	WHERE Timestamp BETWEEN '` + begin_timestamp.Format(TIME_FORMAT) + `' AND '` + end_timestamp.Format(TIME_FORMAT) + `'
	`,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	// Iterate on the result and print it
	for rows.Next() {
		var (
			Query_Tr1_Max        int
			Query_Tr1_Mean       float64
			Query_Web_Mean       float64
			Query_Web_Min        int
			Query_Tr3_Max        int
			Query_Tr3_Mean       float64
			Query_Web_Variance   float64
			Query_Width_Mean     float64
			Query_Threshold_Mean float64
		)
		err := rows.Scan(&Query_Tr1_Max, &Query_Tr1_Mean, &Query_Web_Mean, &Query_Web_Min, &Query_Tr3_Max, &Query_Tr3_Mean, &Query_Web_Variance, &Query_Width_Mean, &Query_Threshold_Mean)
		log.Println("\nQuery_Tr1_Max : ", Query_Tr1_Max,
			"\nQuery_Tr1_Mean : ", Query_Tr1_Mean,
			"\nQuery_Web_Mean ", Query_Web_Mean,
			"\nQuery_Web_Min ", Query_Web_Min,
			"\nQuery_Tr3_Max ", Query_Tr3_Max,
			"\nQuery_Tr3_Mean ", Query_Tr3_Mean,
			"\nQuery_Web_Variance ", Query_Web_Variance,
			"\nQuery_Width_Mean ", Query_Width_Mean,
			"\nQuery_Threshold_Mean ", Query_Threshold_Mean,
		)
		if err != nil {
			return err
		}

	}
	if err := rows.Err(); err != nil {
		log.Println("err2")
		return err
	}
	return nil
}
