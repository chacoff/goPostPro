/*
 * File:    klustersDB.go
 * Date:    March 12, 2025
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Support functions related to the handling of the type of timestamps
 *
 */

package klusters

import (
	"fmt"
	"strings"
	"time"
)

// returnUnixTS gets the timestamps from the DB format and returns the timestamp in unix format
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

// parseFlexibleTimestamp returns a timestamp with the proper quantity of milliseconds digits
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
