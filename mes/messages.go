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
	"strconv"
	"time"

	"goPostPro/global"
	"goPostPro/graphic"
	"goPostPro/postpro"
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

// processType return the real values to answer process messages according the number of passes
func processType(_bodyStatic []interface{}, _bodyDynamic []interface{}, lastTimeStamp string) []interface{} {
	var _bodyAns []interface{}

	passCounter := _bodyStatic[4].(uint32) /// pass counter
	beamId := _bodyStatic[0].(uint32)
	graphic.ChangeName(strconv.FormatUint(uint64(beamId), 10))
	listOfStamps := parseTimeStamps(passCounter, _bodyDynamic, lastTimeStamp)
	log.Println("[MES PostPro] process timestamps", listOfStamps)

	_bodyAns = append(_bodyAns, _bodyStatic[0]) // unique product ID
	_bodyAns = append(_bodyAns, _bodyStatic[1]) // rolling campaign profile
	_bodyAns = append(_bodyAns, _bodyStatic[2]) // rolling campaign number
	_bodyAns = append(_bodyAns, _bodyStatic[3]) // roll stand number
	_bodyAns = append(_bodyAns, _bodyStatic[4]) // pass counter

	// passDates are available in positions 1, 4, 7, 10, 13, 16, 19 ... = pass+(i*2)
	for i := 0; i < int(passCounter); i++ {
		newData, err := postpro.DATABASE.Query_database(listOfStamps[i], listOfStamps[i+1])
		if err != nil {
			log.Println("ERROR : ", err)
		}
		newData.PassNumber = uint32(i + 1)
		newData.PassDate = listOfStamps[i] // time.Now().Format("20060102150405"),
		newData.Dummy = "du"

		_bodyAns = append(_bodyAns, newData.PassNumber)
		_bodyAns = append(_bodyAns, newData.PassDate)
		_bodyAns = append(_bodyAns, newData.Dummy)
		_bodyAns = append(_bodyAns, newData.MaxTempMill3)
		_bodyAns = append(_bodyAns, uint32(newData.AvgTempMill3))
		_bodyAns = append(_bodyAns, newData.MaxTempMill1)
		_bodyAns = append(_bodyAns, uint32(newData.AvgTempMill1))
		_bodyAns = append(_bodyAns, newData.MinTempWeb)
		_bodyAns = append(_bodyAns, uint32(newData.AvgTempWeb))
		_bodyAns = append(_bodyAns, uint32(newData.AvgStdTemp))
		_bodyAns = append(_bodyAns, uint32(newData.PixWidth))
	}

	log.Println("[MES PostPro] post-pro answer", _bodyAns)
	return _bodyAns
}

// parseTimeStamps creates a list with all timeStamps
func parseTimeStamps(passCounter uint32, bodyValuesDynamic []interface{}, lastStamp string) []string {
	var listOfStamps []string

	// passDates are available in positions 1, 4, 7, 10, 13, 16, 19 ... = pass+(i*2)
	for i := 0; i < int(passCounter); i++ {
		pass := i + 1
		listOfStamps = append(listOfStamps, bodyValuesDynamic[pass+(i*2)].(string))
	}

	listOfStamps = append(listOfStamps, lastStamp)

	return listOfStamps
}
