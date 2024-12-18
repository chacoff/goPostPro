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
	"fmt"
	"goPostPro/global"
	"log"
	"math"
	"strings"
	"time"

	"goPostPro/graphic"
	"image/color"

	_ "github.com/mattn/go-sqlite3"
)

var insert_since_cleaning int = 0

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

type CalculationsDatabase struct {
	database *sql.DB
}

// StartDatabase starts db at init of the software
func StartDatabase() error {

	DATABASE = CalculationsDatabase{}

	opening_error := DATABASE.openDatabase()

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

// openDatabase opens the DB while the software init itself
func (calculationsDatabase *CalculationsDatabase) openDatabase() error {

	database, openingError := sql.Open("sqlite3", global.DBParams.Path)

	if openingError != nil {
		return openingError
	}

	calculationsDatabase.database = database

	return nil
}

func (calculationsDatabase *CalculationsDatabase) create_Table() error {

	_, queryError := calculationsDatabase.database.Exec(`
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
			Filename  TEXT,
			ProcessID TEXT,
			Treated   INTEGER CHECK (Treated IN (0, 1)),
			Moving	  INTEGER CHECK (Treated IN (0, 1))
		);`)

	return queryError
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func (calculationsDatabase *CalculationsDatabase) dropTable() error {

	_, queryError := calculationsDatabase.database.Exec(`DROP TABLE IF EXISTS Measures;`)

	return queryError
}

func (calculationsDatabase *CalculationsDatabase) Insert_line_processing(line LineProcessing) error {
	// TODO add sheetpile ID in the DB

	preparation, preparation_error := calculationsDatabase.database.Prepare(
		"INSERT INTO Measures(Timestamp, Tr1_Max, Tr1_Mean, Web_Mean, Web_Min, Tr3_Max, Tr3_Mean, Width, Threshold, Filename, ProcessID, Treated, Moving) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	)

	if preparation_error != nil {
		return preparation_error
	}
	defer preparation.Close()

	// Execute it with the given values
	_, executionError := preparation.Exec(line.timestamp.Format(global.PostProParams.TimeFormat),
		int64(line.max_Tr1),
		int64(line.mean_Tr1),
		int64(line.mean_Web),
		int64(line.min_Web),
		int64(line.max_Tr3),
		int64(line.mean_Tr3),
		int64(line.width),
		int64(line.threshold),
		line.filename,
		fmt.Sprint(global.ProcessID),
		0,
		line.isMoving)
	if executionError != nil {
		return executionError
	}

	// Clean the database every x insertions
	insert_since_cleaning++
	_ = calculationsDatabase.callCleanTable()

	return nil
}

// callCleanTable calls the function to clean the DB and handle the associated errors
func (calculationsDatabase *CalculationsDatabase) callCleanTable() error {

	if insert_since_cleaning >= global.DBParams.CleaningPeriod {

		cleaningError := calculationsDatabase.cleanTable()

		if cleaningError != nil {
			log.Println("[DATABASE] Cleaning error:", cleaningError)
			return cleaningError
		}

		log.Println("[DATABASE] Cleaned")
		insert_since_cleaning = 0
	}

	return nil
}

// cleanTable cleans the DB after certain quantity of lines all the data before certain period of time to keep it small with the current exchange data only
func (calculationsDatabase *CalculationsDatabase) cleanTable() error {

	limitTimestamp := time.Now().Add(time.Duration(-global.DBParams.CleaningHoursKept) * time.Hour)

	_, queryError := calculationsDatabase.database.Exec(`DELETE FROM Measures WHERE Timestamp<'` + limitTimestamp.Format(global.PostProParams.TimeFormat) + `';`)

	return queryError
}

// QueryDatabase will fetch data from the database to calculate the post-processing information
func (calculationsDatabase *CalculationsDatabase) QueryDatabase(begin_string_timestamp string, end_string_timestamp string, pass int) (PostProData, error) {

	passF := passFormater(pass)

	log.Printf("[DATABASE] Processing pass: %s for process ID %d", passF, global.ProcessID)

	post_pro_data := PostProData{}

	begin_timestamp, parsing_error := time.Parse(global.DBParams.TimeFormatRequest, begin_string_timestamp)

	if parsing_error != nil {
		return post_pro_data, parsing_error
	}

	end_timestamp, parsing_error := time.Parse(global.DBParams.TimeFormatRequest, end_string_timestamp)

	if parsing_error != nil {
		return post_pro_data, parsing_error
	}

	graphic.DrawHLineAtTimestamp(begin_string_timestamp, color.RGBA64{255, 0, 0, 255})
	graphic.DrawHLineAtTimestamp(end_string_timestamp, color.RGBA64{0, 255, 0, 255})

	rows, query_error := calculationsDatabase.database.Query(`
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
		WHERE Timestamp BETWEEN ? AND ?
		AND Treated = 0
		AND Filename = ?
		AND ProcessID = ?)
	WHERE Timestamp BETWEEN ? AND ?
	AND Treated = 0
	AND Filename = ?
	AND ProcessID = ?
	`,
		begin_timestamp.Format(global.PostProParams.TimeFormat),
		end_timestamp.Format(global.PostProParams.TimeFormat),
		passF,
		global.ProcessID,
		begin_timestamp.Format(global.PostProParams.TimeFormat),
		end_timestamp.Format(global.PostProParams.TimeFormat),
		passF,
		global.ProcessID)

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

	post_pro_data.AvgStdTemp = math.Sqrt(post_pro_data.AvgStdTemp)

	if row_error := rows.Err(); row_error != nil {
		return post_pro_data, row_error
	}

	return post_pro_data, nil
}

// passFormater formats the pass as a string to be used in the DB-query
func passFormater(pass int) string {

	var passF string

	if global.PostProParams.Cage12Split {
		passF = "Pass undefined"
	} else {
		passF = fmt.Sprintf("Pass %d", pass+1)
	}

	return passF
}

// FindLTCRow finds the LTC row in within the timestamps of the passes
func (calculationsDatabase *CalculationsDatabase) FindLTCRow(begin_string_timestamp string, end_string_timestamp string, pass int) string {
	// @jaime: pay attention the 'pass' here is actually the counter of the passes to process, the counter starts from 0, that's why later is pass+1
	time.Sleep(5 * time.Millisecond)

	passF := passFormater(pass)

	begin_timestamp, _ := time.Parse(global.DBParams.TimeFormatRequest, begin_string_timestamp)
	end_timestamp, _ := time.Parse(global.DBParams.TimeFormatRequest, end_string_timestamp)

	log.Printf("[LTC] Processing LTC %s for process ID %d between %s - %s", passF, global.ProcessID, begin_string_timestamp, end_string_timestamp)

	var timestampLTC string
	err := calculationsDatabase.database.QueryRow(`
		SELECT Timestamp FROM Measures
		WHERE Moving = 1 AND Filename = ? AND Timestamp BETWEEN ? AND ? AND ProcessID = ?
		LIMIT 1
		`,
		passF,
		begin_timestamp.Format(global.PostProParams.TimeFormat),
		end_timestamp.Format(global.PostProParams.TimeFormat),
		global.ProcessID).Scan(&timestampLTC)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("[LTC] No matching LTC timestamp found")
		} else {
			log.Println("[LTC] Error querying database:", err)
		}
	}

	// log.Println("LTC Timestamp:", timestampLTC)
	log.Printf("[LTC] Found the following LTC timestamp reference: %s", timestampLTC)
	formattedTimestampLTC, errFormat := formatTimestamp(timestampLTC)

	if errFormat != nil {
		log.Printf("[LTC] error formating LTC timestamp: %s. Error: %s ", timestampLTC, errFormat)
	}

	log.Printf("[LTC] Processing LTC Timestamp: %s for pass %s for process ID %d", formattedTimestampLTC, passF, global.ProcessID)

	return formattedTimestampLTC
}

// formatTimeStamp bug fix while parsing the timestamps between strings and time.Time types
func formatTimestamp(input string) (string, error) {
	// Parse the input string

	input = padMilliseconds(input)

	t, err := time.Parse("2006-01-02 15:04:05,000", input)
	if err != nil {
		return "", err
	}

	// Format to the desired output
	return t.Format(global.DBParams.TimeFormatRequest), nil
}

// padMilliseconds, helper function to pad milliseconds part with zeros or add ",000" if missing
func padMilliseconds(input string) string {

	if strings.Contains(input, ",") { // Check if the input contains a comma (indicating milliseconds)
		parts := strings.Split(input, ",")
		if len(parts) != 2 {
			return input
		}

		ms := parts[1] // Pad the milliseconds part with zeros to ensure it has exactly 3 digits
		for len(ms) < 3 {
			ms += "0"
		}

		return parts[0] + "," + ms
	}

	// If there's no milliseconds part, add ".000"
	return input + ",000"
}

// FindLTCrealized finds the LTC temperature of the sheetpile about to exist the cage according the pass
// lint:ignore U1000 Ignore unused function temporarily for debugging
func (calculationsDatabase *CalculationsDatabase) FindLTCrealized(begin_string_timestamp string, end_string_timestamp string, pass int) uint32 {
	// 'pass' here is the actual pass from the LTC message

	begin_timestamp, _ := time.Parse(global.DBParams.TimeFormatRequest, begin_string_timestamp)
	end_timestamp, _ := time.Parse(global.DBParams.TimeFormatRequest, end_string_timestamp)

	passF := fmt.Sprintf("Pass %d", pass)
	log.Printf("[DATABASE] Processing LTC-realized for pass: %s for process ID %d", passF, global.ProcessID)

	var LTCrealized uint32
	err := calculationsDatabase.database.QueryRow(`
		SELECT Tr3_Max FROM Measures
		WHERE Moving = 1 
		AND Filename = ?
		AND Timestamp BETWEEN ? AND ?
		LIMIT 1
		`,
		passF,
		begin_timestamp.Format(global.PostProParams.TimeFormat),
		end_timestamp.Format(global.PostProParams.TimeFormat)).Scan(&LTCrealized)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("[LTC] No matching LTC timestamp found")
		} else {
			log.Println("[LTC] Error querying database:", err)
		}
	}

	log.Printf("[DATABASE] Processed LTC-realized: %d for pass %s for process ID %d", LTCrealized, passF, global.ProcessID)

	return LTCrealized
}

// updateTreated updates all the treated rows with a 1 to avoid include them in future post-processing
// lint:ignore U1000 Ignore unused function temporarily for debugging
func (calculationsDatabase *CalculationsDatabase) UpdateTreated(beginStr string, endStr string) (int64, error) {

	begin, _ := time.Parse(global.DBParams.TimeFormatRequest, beginStr)
	end, _ := time.Parse(global.DBParams.TimeFormatRequest, endStr)

	query := `
    UPDATE Measures
    SET Treated = 1
    WHERE Timestamp BETWEEN ? AND ?
    `

	result, err := calculationsDatabase.database.Exec(query,
		begin.Format(global.PostProParams.TimeFormat),
		end.Format(global.PostProParams.TimeFormat))

	if err != nil {
		log.Println("[DATABASE] error updating Traited status:", err)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("[DATABASE] error getting rows affected:", err)
		return 0, err
	}

	return rowsAffected, nil
}
