package main

import (
	"encoding/binary"
	"fmt"
)

func encodeUint32(_values []interface{}) []byte {
	buffer := make([]byte, 0, MaxBufferSize) // variable buffer with maximum capacity MaxBufferSize

	j := 0 // Little and Big endian are mixed - only identification in big, the rest in little
	for _, value := range _values {
		binaryValue := make([]byte, 4)
		if j == 1 { // identification is the element 1 in the slice_headerType
			binary.BigEndian.PutUint32(binaryValue, value.(uint32))
		} else {
			binary.LittleEndian.PutUint32(binaryValue, value.(uint32))
		}
		buffer = append(buffer, binaryValue...) // slice, element to unpack
		j += 1
	}

	if verbose {
		fmt.Printf("Length of buffer: %d\n", len(buffer))
		fmt.Printf("Capacity of buffer: %d\n", cap(buffer))
	}

	return buffer
}

func encodeProcess(_values []interface{}) []byte {
	buffer := make([]byte, 0, MaxBufferSize)

	return buffer
}
