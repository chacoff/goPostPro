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

func decodeHeaderUint32(data []byte) []uint32 {
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

	return _values
}

func decodeBody(data []byte) []uint32 {
	var _valuesStatic []uint32

	// fmt.Printf(">> Received body (hex): %s\n", hex.EncodeToString(data))

	bodyStatic := data[:staticBodySize]
	//bodyDynamic := data[staticBodySize:]

	_valuesStatic = decodeBodyStatic(bodyStatic)

	return _valuesStatic
}

func decodeBodyStatic(data []byte) []uint32 {
	// 0 uint32
	// 1 utf
	// 2 utf
	// 3 uint32
	// 4 uint32

	var _values []uint32
	var value uint32

	for i := 0; i+1 < len(data); i += bytesGap {
		if i == 0 || i == 3 || i == 4 {
			value = binary.LittleEndian.Uint32(data[i : i+bytesGap])
		} else {
			hexBytes, _ := hex.DecodeString(hex.EncodeToString(data[i : i+bytesGap]))
			utf8String := string(hexBytes)
			fmt.Println(utf8String)
		}

		if verbose {
			fmt.Printf("-- %s - %d decoded: %d\n", hex.EncodeToString(data[i:i+bytesGap]), i, value)
		}
		_values = append(_values, value)
	}
	return _values
}
