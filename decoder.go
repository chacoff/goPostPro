/*
 * File:    decoder.go
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
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// global variables specific for decoding
const (
	bytesGap       int = 4
	staticBodySize int = 20
)

func decodeHeaderUint32(data []byte) ([]uint32, bool) {
	// Total Length			format	HEX		BytesGap	4
	// Identification		format	HEX		BytesGap	4
	// Message Counter		format	HEX		BytesGap	4
	// Year					format	HEX		BytesGap	4
	// Month				format	HEX		BytesGap	4
	// Day					format	HEX		BytesGap	4
	// Hour					format	HEX		BytesGap	4
	// Minute				format	HEX		BytesGap	4
	// Second				format	HEX		BytesGap	4
	// Hundred-of-Seconds	format	HEX		BytesGap	4

	var _values []uint32
	var value uint32
	j := 0
	length := binary.BigEndian.Uint32(data[0:bytesGap]) // it assumes BigEndian

	// fmt.Printf(">> Received header (hex): %s\n", hex.EncodeToString(data))

	for i := 0; i+1 < len(data); i += bytesGap { // iterate over the byte slice in steps of 4 bytes, i.e. 8 characters
		if length > 1000 && j < 3 { // because watchdog comes in big endian and process messages in both!!!
			value = binary.LittleEndian.Uint32(data[i : i+bytesGap])
		} else {
			value = binary.BigEndian.Uint32(data[i : i+bytesGap])
		}

		if verbose {
			fmt.Printf("-- %s - %d decoded: %d\n", hex.EncodeToString(data[i:i+bytesGap]), i, value)
		}
		j += 1
		_values = append(_values, value)
	}

	return _values, true
}

func decodeBody(data []byte, messageType int) ([]interface{}, []interface{}) {
	var _valuesStatic []interface{}
	var _valuesDynamic []interface{}

	if messageType == 4704 || messageType == 4714 { // it is LTC message
		_valuesStatic = decodeLTC(data)
	} else {
		bodyStatic := data[:staticBodySize]
		bodyDynamic := data[staticBodySize:]
		_valuesStatic = decodeBodyStatic(bodyStatic)
		nPasses := int(_valuesStatic[4].(uint32))
		_valuesDynamic = decodePasses(bodyDynamic, nPasses) // dynamic data in bytes and number of passes
	}
	return _valuesStatic, _valuesDynamic
}

func decodeBodyStatic(data []byte) []interface{} {
	// j=0 Unique ID 		format	HEX		BytesGap	4
	// j=1 Roll Profile 	format 	UTF8	BytesGap	4
	// j=2 Roll Number 		format 	UTF8	BytesGap	4
	// j=3 Roll Stand 		format 	HEX		BytesGap	4
	// j=4 Pass Counter 	format 	HEX		BytesGap	4

	var _values []interface{}
	var value uint32
	var valueUtf string

	j := 0
	for i := 0; i+1 < len(data); i += bytesGap {
		if j == 0 || j == 3 || j == 4 {
			value = binary.LittleEndian.Uint32(data[i : i+bytesGap])
			_values = append(_values, value)
		} else {
			hexBytes, _ := hex.DecodeString(hex.EncodeToString(data[i : i+bytesGap]))
			valueUtf = string(hexBytes)
			_values = append(_values, valueUtf)
		}

		if verbose {
			if j == 0 || j == 3 || j == 4 {
				fmt.Printf("-- %s - %d decoded: %d\n", hex.EncodeToString(data[i:i+bytesGap]), i, value)
			} else {
				fmt.Printf("-- %s - %d decoded: %s\n", hex.EncodeToString(data[i:i+bytesGap]), i, valueUtf)
			}
		}
		j += 1
	}
	return _values
}

func decodePasses(data []byte, passes int) []interface{} {
	// Pass Number	format	HEX			BytesGap	4
	// Pass Date	format 	timestamp	BytesGap	14
	// Dummy		format 	HEX			BytesGap	2

	var byteGaps = []int{4, 14, 2} // every pass is 20bytes: pattern of byte gaps to decode specific messages
	var index int
	var _values []interface{}
	var value uint32
	var timestamp string
	var dummy uint16

	for i := 1; i <= passes; i++ { // i acts as pass numbers
		for _, gap := range byteGaps {
			endIndex := index + gap
			if endIndex > len(data) {
				endIndex = len(data)
			}
			_data := data[index:endIndex] // Extract bytes according to the pattern

			switch {
			case gap == 4: // pass number
				value = binary.LittleEndian.Uint32(_data)
				_values = append(_values, value)
			case gap == 14: // pass date
				hexBytes, _ := hex.DecodeString(hex.EncodeToString(_data))
				timestamp = string(hexBytes)
				_values = append(_values, timestamp)
			case gap == 2: // dummy
				dummy = binary.LittleEndian.Uint16(_data)
				_values = append(_values, dummy)
			}
			index = endIndex
		}
	}
	return _values
}

func decodeLTC(data []byte) []interface{} {
	// j=0	id_482 			format HEX	BytesGap	4
	// j=1 	grp_mont_482 	format UTF 	BytesGap	4
	// j=2	num_mont_482 	format UTF	BytesGap	4
	// j=3	cage_482 		format HEX	BytesGap	4
	// j=4	code_prod_482 	format UTF	BytesGap	12
	// j=5	nuan_train_482 	format UTF	BytesGap	7
	// j=6	dummy_482 		format UTF	BytesGap	1
	// j=7	temp_ltc_482	format HEX	BytesGap	4

	var _values []interface{}
	var byteGaps = []int{4, 4, 4, 4, 12, 7, 1, 4} // pattern of byte gaps to decode specific messages
	var index int
	var value uint32

	j := 0
	for _, gap := range byteGaps {
		endIndex := index + gap
		if endIndex > len(data) {
			endIndex = len(data)
		}
		_data := data[index:endIndex] // Extract bytes according to the pattern

		if j == 0 || j == 3 || j == 7 {
			value = binary.LittleEndian.Uint32(_data)
			_values = append(_values, value)
		} else {
			hexBytes, _ := hex.DecodeString(hex.EncodeToString(_data))
			_values = append(_values, string(hexBytes))
		}

		index = endIndex
		j += 1
	}

	return _values
}
