/*
 * File:    database.go
 * Date:    May 10, 2024
 * Author:  T.V
 * Email:   theo.verbrugge77@gmail.com
 * Project: goPostPro
 * Description:
 *   Contains the functions and queries to use the sqlite database
 *
 */

package postpro

import (
	"database/sql"
	"goPostPro/global"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type PostProData struct {
	PassNumber   uint32
	PassDate     string
	Dummy        string
	MaxTempMill3 uint32
	AvgTempMill3 float64
	MaxTempMill1 uint32
	AvgTempMill1 float64
	MinTempWeb   uint32
	AvgTempWeb   float64
	AvgStdTemp   float64
	PixWidth     float64
}

func StartDatabase() error {
	DATABASE = CalculationsDatabase{}
	opening_error := DATABASE.open_database()
	if opening_error != nil {
		return opening_error
	}
	// drop_error := DATABASE.drop_Table()
	// if drop_error != nil {
	// 	return drop_error
	// }
	creation_error := DATABASE.create_Table()
	if creation_error != nil {
		return creation_error
	}
	log.Println("[DATABASE] init with success")
	return nil
}

type CalculationsDatabase struct {
	database *sql.DB
}

func (calculations_database *CalculationsDatabase) open_database() error {
	database, opening_error := sql.Open("sqlite3", global.DBParams.Path)
	if opening_error != nil {
		return opening_error
	}
	calculations_database.database = database
	return nil
}

func (calculations_database *CalculationsDatabase) create_Table() error {
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

func (calculations_database *CalculationsDatabase) drop_Table() error {
	_, query_error := calculations_database.database.Exec(`DROP TABLE IF EXISTS Measures;`)
	return query_error
}

func (calculations_database *CalculationsDatabase) Insert_line_processing(line LineProcessing) error {
	preparation, preparation_error := calculations_database.database.Prepare(
		"INSERT INTO Measures(Timestamp, Tr1_Max, Tr1_Mean, Web_Mean, Web_Min, Tr3_Max, Tr3_Mean, Width, Threshold, Filename) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if preparation_error != nil {
		return preparation_error
	}
	defer preparation.Close()
	// Execute it with the given values
	_, execution_error := preparation.Exec(line.timestamp.Format(global.PostProParams.TimeFormat), int64(line.max_Tr1), int64(line.mean_Tr1), int64(line.mean_Web), int64(line.min_Web), int64(line.max_Tr3), int64(line.mean_Tr3), int64(line.width), int64(line.threshold), line.filename)
	if execution_error != nil {
		return execution_error
	}
	return nil
}

func (calculations_database *CalculationsDatabase) Query_database(begin_string_timestamp string, end_string_timestamp string) (PostProData, error) {
	post_pro_data := PostProData{}
	begin_timestamp, parsing_error := time.Parse(global.DBParams.TimeFormatRequest, begin_string_timestamp)
	if parsing_error != nil {
		return post_pro_data, parsing_error
	}
	end_timestamp, parsing_error := time.Parse(global.DBParams.TimeFormatRequest, end_string_timestamp)
	if parsing_error != nil {
		return post_pro_data, parsing_error
	}

	rows, query_error := calculations_database.database.Query(`
	SELECT
		COALESCE(MAX(Tr1_Max), 0) AS Query_Tr1_Max,
		COALESCE(AVG(Tr1_Mean), 0) AS Query_Tr1_Mean,
		COALESCE(Query_Web_Mean, 0),
		COALESCE(MIN(Web_Min), 0) AS Query_Web_Min,
		COALESCE(MAX(Tr3_Max), 0) AS Query_Tr3_Max,
		COALESCE(AVG(Tr3_Mean), 0) AS Query_Tr3_Mean,
		COALESCE(AVG((Web_Mean-Query_Web_Mean)*(Web_Mean-Query_Web_Mean)), 0) AS Query_Web_Variance,
		COALESCE(AVG(Width), 0) AS Query_Width_Mean,
		COALESCE(AVG(Threshold), 0) AS Query_Threshold_Mean
	FROM Measures,
		(SELECT AVG(Web_Mean) AS Query_Web_Mean
		FROM Measures
		WHERE Timestamp BETWEEN '` + begin_timestamp.Format(global.PostProParams.TimeFormat) + `' AND '` + end_timestamp.Format(global.PostProParams.TimeFormat) + `')
	WHERE Timestamp BETWEEN '` + begin_timestamp.Format(global.PostProParams.TimeFormat) + `' AND '` + end_timestamp.Format(global.PostProParams.TimeFormat) + `'
	`,
	)
	if query_error != nil {
		return post_pro_data, query_error
	}
	defer rows.Close()
	// Iterate on the result and print it
	for rows.Next() {
		Query_Threshold_Mean := float64(0)
		scan_error := rows.Scan(
			&post_pro_data.MaxTempMill1,
			&post_pro_data.AvgTempMill1,
			&post_pro_data.AvgTempWeb,
			&post_pro_data.MinTempWeb,
			&post_pro_data.MaxTempMill3,
			&post_pro_data.AvgTempMill3,
			&post_pro_data.AvgStdTemp,
			&post_pro_data.PixWidth,
			&Query_Threshold_Mean)

		if scan_error != nil {
			return post_pro_data, scan_error
		}

	}
	if row_error := rows.Err(); row_error != nil {
		return post_pro_data, row_error
	}
	return post_pro_data, nil
}
