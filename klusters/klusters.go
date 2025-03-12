/*
 * File:    klusters.go
 * Date:    March 09, 2025
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Using K-means clusters all measures in order to re-assign the Pass number
 *
 */

package klusters

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"goPostPro/global"
	"goPostPro/postpro"
	"strings"
	"time"

	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

// getDB gets an instance of the DB from postpro package
func getDB() (*sql.DB, error) {
	db := postpro.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	return db, nil
}

// getData function is used to get all the data between timestamps prior clustering the new passes
func getData(beginTS string, endTS string, beamID uint32) ([]string, error) {

	beginTS_format, _ := time.Parse(global.DBParams.TimeFormatRequest, beginTS)
	endTS_format, _ := time.Parse(global.DBParams.TimeFormatRequest, endTS)

	var timeStamps []string
	db, _ := getDB()

	sqlQuery := `SELECT Timestamp, Filename FROM Measures WHERE Timestamp BETWEEN ? AND ? AND ProcessID = ?`
	rows, queryError := db.Query(sqlQuery, beginTS_format, endTS_format, beamID)

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

// updatePass updates the pass according the cluster result
func updatePass(beginTS string, endTS string, pass string, beamID uint32) error {

	db, _ := getDB()
	fmt.Printf("Updating %s between %s and %s for Beam %d\n", pass, beginTS, endTS, beamID)

	sqlQuery := `UPDATE Measures SET Cluster = ? WHERE Timestamp BETWEEN ? AND ? AND ProcessID = ?`
	_, err := db.Exec(sqlQuery, pass, beginTS, endTS, beamID)
	if err != nil {
		return err
	}

	return nil
}

// ReClusterPasses will re-assign the pass by using kmeans over the timestamps of the measures
func ReClusterPasses(beginTS string, endTS string, k int, beamID uint32) {
	var firstCoordinates clusters.Coordinates
	var lastCoordinates clusters.Coordinates

	timestamps, errorGetData := getData(beginTS, endTS, beamID)

	if errorGetData != nil {
		fmt.Println("Error getting data from DB to start clustering")
		return
	}
	unixTimes, timestampMap := returnUnixTS(timestamps)

	// Clustering, convert 1D data to clusters.Observations
	var observations clusters.Observations

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

		if len(c.Observations) > 0 {
			firstObs := c.Observations[0]
			lastObs := c.Observations[len(c.Observations)-1]

			firstCoordinates = firstObs.(clusters.Coordinates)
			lastCoordinates = lastObs.(clusters.Coordinates)
		}

		pass := fmt.Sprintf("Pass %d", i+1)
		errorUpdate := updatePass(timestampMap[firstCoordinates[0]], timestampMap[lastCoordinates[0]], pass, beamID)
		if errorUpdate != nil {
			return
		}
	}
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
