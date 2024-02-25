package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func decodeHeaderUint32(data []byte) []uint32 {
	var _values []uint32
	var value uint32
	j := 0
	length := binary.BigEndian.Uint32(data[0:4]) // it assumes BigEndian

	for i := 0; i+1 < len(data); i += 4 { // iterate over the byte slice in steps of 4 bytes, i.e. 8 characters
		if length > 1000 && j < 3 { // because watchdog comes in big endian and process messages in both!!!
			value = binary.LittleEndian.Uint32(data[i : i+4])
		} else {
			value = binary.BigEndian.Uint32(data[i : i+4])
		}

		if verbose {
			fmt.Printf("-- %s - %d decoded: %d\n", hex.EncodeToString(data[i:i+4]), i, value)
		}
		j += 1
		_values = append(_values, value)
	}

	return _values
}
