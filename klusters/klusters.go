package klusters

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
	"time"

	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

type db struct {
	database *sql.DB
}

var dataBase db = db{}

func klustering() {

	timestamps := returnTS()
	unixTimes, timestampMap := returnUnixTS(timestamps)

	// Clustering, convert 1D data to clusters.Observations
	var observations clusters.Observations
	var k int = 3

	for _, value := range unixTimes {
		observations = append(observations, clusters.Coordinates{value})
	}

	km := kmeans.New()
	clustersPasses, err := km.Partition(observations, k)
	if err != nil {
		fmt.Println("Error partitioning data: ", err)
	}

	// Output results
	for i, c := range clustersPasses {
		fmt.Printf("Cluster %d:\n", i+1)
		//for _, obs := range c.Observations {
		//	coordinates := obs.(clusters.Coordinates)
		//	fmt.Printf("  %.2f - %s\n", coordinates[0], timestampMap[coordinates[0]])
		//}

		if len(c.Observations) > 0 {
			firstObs := c.Observations[0]
			lastObs := c.Observations[len(c.Observations)-1]

			firstCoordinates := firstObs.(clusters.Coordinates)
			lastCoordinates := lastObs.(clusters.Coordinates)

			fmt.Printf("  First: %.2f - %s\n", firstCoordinates[0], timestampMap[firstCoordinates[0]])
			fmt.Printf("  Last: %.2f - %s\n", lastCoordinates[0], timestampMap[lastCoordinates[0]])
		}

		fmt.Printf("Centered at: %.3f\n", c.Center[0])
	}
}

func returnTS() []string {
	dataBase = db{}
	openError := dataBase.openDatabase()
	if openError != nil {
		return nil
	}

	timestamps, queryError := dataBase.getData()
	if queryError != nil {
		fmt.Printf("Error querying database: %v\n", queryError)
		return nil
	}

	return timestamps
}

func returnUnixTS(ts []string) ([]float64, map[float64]string) {
	var unixTimes []float64
	var timestampMap = make(map[float64]string)

	for _, ts := range ts {
		t, err := parseFlexibleTimestamp(ts)
		if err != nil {
			fmt.Printf("Error parsing timestamp: %v\n", err)
			return nil, nil
		}

		unixTime := float64(t.UnixMilli()) // maybe uni nano?
		unixTimes = append(unixTimes, unixTime)
		timestampMap[unixTime] = ts // dict unixTime as key and value timestamp in string
	}

	return unixTimes, timestampMap
}

func parseFlexibleTimestamp(ts string) (time.Time, error) {
	ts = strings.Replace(ts, ",", ".", 1) // Replace , with . due to Go parsing

	layouts := []string{
		"2006-01-02 15:04:05",     // no decimals
		"2006-01-02 15:04:05.0",   // 1 decimal
		"2006-01-02 15:04:05.00",  // 2 decimals
		"2006-01-02 15:04:05.000", // 3 decimals
	}

	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, ts)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, lastErr
}

func (db *db) openDatabase() error {
	database, openingError := sql.Open("sqlite3", "processed.db")

	if openingError != nil {
		return openingError
	}

	db.database = database
	return nil
}

func (db *db) getData() ([]string, error) {
	var timeStamps []string

	sqlQuery, _ := loadSQLFile("sql\\getData.sql")
	rows, queryError := db.database.Query(sqlQuery)

	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	for rows.Next() {
		var timeStamp, fileName string
		scanError := rows.Scan(&timeStamp, &fileName)
		if scanError != nil {
			return nil, scanError
		}
		timeStamps = append(timeStamps, timeStamp)
	}

	if rowError := rows.Err(); rowError != nil {
		return nil, rowError
	}

	return timeStamps, nil
}

func loadSQLFile(filePath string) (string, error) {

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
