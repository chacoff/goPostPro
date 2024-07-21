/*
 * File:    encode.go
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
	"encoding/binary"
	"log"

	"goPostPro/global"
)

// encodeUint32 function encoding Header data, it is common for all message types
func encodeUint32(_values []interface{}) []byte {
	_buffer := make([]byte, 0, global.AppParams.MaxBufferSize) // variable buffer with maximum capacity MaxBufferSize

	j := 0 // Little and Big endian are mixed - only identification in big, the rest in little
	for _, value := range _values {
		binaryValue := make([]byte, 4)
		if j == 1 { // identification is the element 1 in the slice_headerType
			binary.BigEndian.PutUint32(binaryValue, value.(uint32))
		} else {
			binary.LittleEndian.PutUint32(binaryValue, value.(uint32))
		}
		_buffer = append(_buffer, binaryValue...) // slice, element to unpack
		j += 1
	}

	if global.AppParams.Verbose {
		log.Printf("Length of buffer: %d\n", len(_buffer))
		log.Printf("Capacity of buffer: %d\n", cap(_buffer))
	}

	return _buffer
}

// encodeProcess encode process data containing passes, it is only to encode Body
func encodeProcess(_values []interface{}) []byte {
	// some of the data has to be encoded Raw, i.e., CHARS to BYTE and others using littleEndian to UINT32

	_buffer := make([]byte, 0, global.AppParams.MaxBufferSize)
	var rawWrite []int

	// new Protocol is already integrated in MES
	rawWrite = []int{1, 2, 6, 7, 23, 24, 40, 41, 57, 58, 74, 75, 91, 92, 108, 109, 125, 126, 142, 143, 159, 160, 176, 177, 193, 194, 210, 211}
	// rawWrite = []int{1, 2, 6, 7, 17, 18, 28, 29, 39, 40, 50, 51, 61, 62, 72, 73, 83, 84, 94, 95}

	j := 0 // Little and Big endian are mixed - only identification in big, the rest in little
	for _, value := range _values {

		binaryValue := make([]byte, 4)

		if isInSlice(rawWrite, j) {
			// log.Println(value)
			binaryValue = []byte(value.(string))
		} else {
			binary.LittleEndian.PutUint32(binaryValue, value.(uint32))
		}

		_buffer = append(_buffer, binaryValue...) // slice, element to unpack
		j += 1
	}

	return _buffer
}

func isInSlice(slice []int, elem int) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}
