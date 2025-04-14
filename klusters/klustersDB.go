/*
 * File:    klustersDB.go
 * Date:    March 12, 2025
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Support functions related to the connection with the DB
 *
 */

package klusters

import (
	"database/sql"
	"errors"
	"fmt"
	"goPostPro/global"
	"goPostPro/postpro"
	"time"
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
	fmt.Printf("-- Updating %s between %s and %s for Beam %d\n", pass, beginTS, endTS, beamID)

	sqlQuery := `UPDATE Measures SET Cluster = ? WHERE Timestamp BETWEEN ? AND ? AND ProcessID = ?`
	_, err := db.Exec(sqlQuery, pass, beginTS, endTS, beamID)
	if err != nil {
		return err
	}

	return nil
}
