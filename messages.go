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

package main

import (
	"fmt"
	"time"
)

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

	if verbose {
		fmt.Println("Header to encode:", _values)
	}

	return _values
}

func processType(_id uint32, _counter uint32, _bodyStatic []interface{}, _bodyDynamic []interface{}) []interface{} {
	var _values []interface{}
	// var _header []interface{}

	_values = append(_values, _bodyStatic[0]) // unique product ID
	_values = append(_values, _bodyStatic[1]) // rolling campaign profile
	_values = append(_values, _bodyStatic[2]) // rolling campaign number
	_values = append(_values, _bodyStatic[3]) // roll stand number
	_values = append(_values, _bodyStatic[4]) // pass counter

	// _header = headerType(40, _id, _counter) // add the length after
	return _values
}
