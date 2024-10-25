/*
 * File:    messages.go
 * Date:    March 04, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Gathers data from thermal cameras at Train2 and cross-match with timestamps coming from MES to
 *	 to outcome post processes data.
 */

package mes

import (
	"log"
	"time"

	"goPostPro/global"
)

// headerType return the values of header, at this stage nothing is encoded, it is a vector with the real values
func headerType(_size uint32, _id uint32, _counter uint32) []interface{} {
	var _values []interface{} // interface, even though all are uint32 due to body being interface{}

	_now := time.Now()

	_values = append(_values, _size)                              // total length, always 40 in header
	_values = append(_values, _id)                                // identification
	_values = append(_values, _counter)                           // message counter
	_values = append(_values, uint32(_now.Year()))                // year
	_values = append(_values, uint32(_now.Month()))               // month
	_values = append(_values, uint32(_now.Day()))                 // day
	_values = append(_values, uint32(_now.Hour()))                // hours
	_values = append(_values, uint32(_now.Minute()))              // minutes
	_values = append(_values, uint32(_now.Second()))              // seconds
	_values = append(_values, uint32(_now.Nanosecond()/10000000)) // hundreds of seconds

	if global.AppParams.Verbose {
		log.Println("[MES Header] Header to encode:", _values)
	}

	return _values
}

// parseTimeStamps creates a list with all timeStamps
func parseTimeStamps(passCounter uint32, bodyValuesDynamic []interface{}, lastStamp string) []string {
	// from bodyValuesDynamic, passDates are available in positions 1, 4, 7, 10, 13, 16, 19 ... = pass+(i*2)

	var listOfStamps []string

	for i := 0; i < int(passCounter); i++ {
		pass := i + 1
		listOfStamps = append(listOfStamps, bodyValuesDynamic[pass+(i*2)].(string))
	}

	listOfStamps = append(listOfStamps, lastStamp)

	return listOfStamps
}

// addOffsetToTimestamp adds offset in timestamp to calculate the instance LTC
func addOffsetToTimestamp(timestamp string, offset int) string {

	log.Printf("[LTC offset] timestamp to parse: %s", timestamp)
	newStamp, parsing_error := time.Parse(global.DBParams.TimeFormatRequest, timestamp)

	if parsing_error != nil {
		log.Println("[LTC offset] parsing error: ", parsing_error)
	}

	newStamp = newStamp.Add(time.Duration(offset) * time.Second)

	return newStamp.Format(global.DBParams.TimeFormatRequest)
}
