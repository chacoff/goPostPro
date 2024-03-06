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

package main

import (
	"encoding/binary"
	"fmt"
)

func encodeUint32(_values []interface{}) []byte {
	_buffer := make([]byte, 0, MaxBufferSize) // variable buffer with maximum capacity MaxBufferSize

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

	if verbose {
		fmt.Printf("Length of buffer: %d\n", len(_buffer))
		fmt.Printf("Capacity of buffer: %d\n", cap(_buffer))
	}

	return _buffer
}

func encodeProcess(_values []interface{}) []byte {
	_buffer := make([]byte, 0, MaxBufferSize)

	var rawWrite = []int{11, 12, 16, 17, 27, 28, 38, 39, 49, 50, 60, 61, 71, 72, 82, 83, 93, 94}

	j := 0 // Little and Big endian are mixed - only identification in big, the rest in little
	for _, value := range _values {

		binaryValue := make([]byte, 4)

		if j == 1 { // identification is the element 1 in the slice_headerType
			binary.BigEndian.PutUint32(binaryValue, value.(uint32))
		} else if isInSlice(rawWrite, j) {
			// fmt.Println(value)
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
