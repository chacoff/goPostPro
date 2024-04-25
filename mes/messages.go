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
	"fmt"
	"time"

	"goPostPro/global"
)

type postProData struct {
	passNumber   uint32
	passDate     string
	dummy        string
	maxTempMill3 uint32
	avgTempMill3 uint32
	maxTempMill1 uint32
	avgTempMill1 uint32
	minTempWeb   uint32
	avgTempWeb   uint32
	avgStdTemp   uint32
	pixWidth     uint32
}

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

	if global.Appconfig.Verbose {
		fmt.Println("Header to encode:", _values)
	}

	return _values
}

// processType return the real values to answer process messages according the number of passes
func processType(_bodyStatic []interface{}, _bodyDynamic []interface{}) []interface{} {
	var _bodyAns []interface{}

	passCounter := _bodyStatic[4].(uint32) /// pass counter

	_bodyAns = append(_bodyAns, _bodyStatic[0]) // unique product ID
	_bodyAns = append(_bodyAns, _bodyStatic[1]) // rolling campaign profile
	_bodyAns = append(_bodyAns, _bodyStatic[2]) // rolling campaign number
	_bodyAns = append(_bodyAns, _bodyStatic[3]) // roll stand number
	_bodyAns = append(_bodyAns, _bodyStatic[4]) // pass counter

	// passDates are available in positions 1, 4, 7, 10, 13, 16, 19 ... = pass+(i*2)
	for i := 0; i < int(passCounter); i++ {
		pass := i + 1
		newData := postProData{
			passNumber:   uint32(pass),
			passDate:     _bodyDynamic[pass+(i*2)].(string), // time.Now().Format("20060102150405"),
			dummy:        "du",
			maxTempMill3: 8,
			avgTempMill3: 0,
			maxTempMill1: 0,
			avgTempMill1: 0,
			minTempWeb:   0,
			avgTempWeb:   0,
			avgStdTemp:   0,
			pixWidth:     5,
		}

		_bodyAns = append(_bodyAns, newData.passNumber)
		_bodyAns = append(_bodyAns, newData.passDate)
		_bodyAns = append(_bodyAns, newData.dummy)
		_bodyAns = append(_bodyAns, newData.maxTempMill3)
		_bodyAns = append(_bodyAns, newData.avgTempMill3)
		_bodyAns = append(_bodyAns, newData.maxTempMill1)
		_bodyAns = append(_bodyAns, newData.avgTempMill1)
		_bodyAns = append(_bodyAns, newData.minTempWeb)
		_bodyAns = append(_bodyAns, newData.avgTempWeb)
		_bodyAns = append(_bodyAns, newData.avgStdTemp)
		_bodyAns = append(_bodyAns, newData.pixWidth)
	}

	return _bodyAns
}
