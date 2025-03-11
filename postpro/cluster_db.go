package postpro

import "fmt"

// GetData function is used to get all the data between timestamps prior clustering the new passes
func (calculationsDatabase *CalculationsDatabase) GetData(beginTS string, endTS string, beamID uint32) ([]string, error) {

	var timeStamps []string

	sqlQuery := `SELECT Timestamp, Filename FROM Measures WHERE Timestamp BETWEEN ? AND ? AND ProcessID = ?`
	rows, queryError := calculationsDatabase.database.Query(sqlQuery, beginTS, endTS, beamID)

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

// UpdatePass updates the pass according the cluster result
func (calculationsDatabase *CalculationsDatabase) UpdatePass(beginTS string, endTS string, pass string, beamID uint32) error {

	fmt.Printf("Updating %s between %s and %s for Beam %d\n", pass, beginTS, endTS, beamID)

	sqlQuery := `UPDATE Measures SET Cluster = ? WHERE Timestamp BETWEEN ? AND ? AND ProcessID = ?`
	_, err := calculationsDatabase.database.Exec(sqlQuery, pass, beginTS, endTS, beamID)
	if err != nil {
		return err
	}

	return nil
}
