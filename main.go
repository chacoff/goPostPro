/*
 * File:    main.go
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
	"net"
	"os"
)

// global variables
const (
	verbose       bool   = false
	netType       string = "tcp"
	address       string = "10.28.114.89:4600" // 10.28.114.89
	MaxBufferSize int    = 2048
	headerSize    int    = 40
)

var buffer = make([]byte, MaxBufferSize)         // buffer to hold incoming data
var allHexBytes = make([]byte, 0, MaxBufferSize) // variable: from 0 to maxBufferSize. Data will be appended on arrival

func main() {
	listener, err := net.Listen(netType, address) // listen on port 4600
	if err != nil {
		fmt.Println("[WARNING] Error listening:", err)
		return
	}

	defer listener.Close() // close the connection when the function returns using a schedule: defer

	fmt.Printf("Server is listening on port %s\n", address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[WARNING] Error accepting connection:", err)
			os.Exit(1)
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	// defer conn.Close()

	var isHeaderOk = false
	var headerValues []uint32
	var FullLength int

	for {
		n, err := conn.Read(buffer) // read data from the connection
		if err != nil {             // && err != io.EOF
			fmt.Println("[WARNING] Error reading or Client disconnected:", err)
			allHexBytes = make([]byte, 0, MaxBufferSize) // reset variable before handling answer
			isHeaderOk = false                           // reset variables before handling answer
			break
		}
		allHexBytes = append(allHexBytes, buffer[:n]...) //  append data upon arrival

		if len(allHexBytes) >= headerSize && isHeaderOk == false {
			hexBytesHeader := allHexBytes[:headerSize]                    // Extract first 40 bytes, only header
			headerValues, isHeaderOk = decodeHeaderUint32(hexBytesHeader) // decode little-endian uint32 values
			FullLength = int(headerValues[0])
			fmt.Println(">> Decoded Header values:", headerValues)
		}

		if len(allHexBytes) >= FullLength && isHeaderOk == true { // TODO attention with the '>='
			hexBytesBody := allHexBytes[headerSize:FullLength] // Extract the rest of the bytes
			allHexBytes = make([]byte, 0, MaxBufferSize)       // reset variable before handling answer
			isHeaderOk = false                                 // reset variables before handling answer
			go handleAnswer(conn, headerValues, hexBytesBody)
		}
	}
}

func handleAnswer(conn net.Conn, _headerValues []uint32, _hexBytesBody []byte) {

	var echo = false
	var response []byte
	messageType := int(_headerValues[1]) // message type on the header
	messageCounter := _headerValues[2]   // already in uint32
	messageTypeAns := uint32(messageType - 100)

	switch messageType {
	case 4701: // watchdog: only header
		response = encodeUint32(headerType(40, messageTypeAns, messageCounter))
		echo = true

	case 4702: // process message: header + body
		bodyValuesStatic, bodyValueDynamic := decodeBody(_hexBytesBody, messageType)
		fmt.Println(">> Decoded Body values:", bodyValuesStatic, bodyValueDynamic)
		// response = encodeProcess(processType(messageTypeAns, messageCounter, bodyValuesStatic, bodyValueDynamic))
		response = []byte("dur process")
		echo = true

	case 4703: // acknowledge data message
		fmt.Println("MES received data properly")
		echo = false

	default:
		fmt.Println("Unknown message:", messageCounter)
		echo = false
	}

	if echo == true {
		go responseWriter(conn, response, messageCounter)
	}
}

func responseWriter(conn net.Conn, _response []byte, _messageCounter uint32) {
	_, err := conn.Write(_response)
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}
	fmt.Println("Response sent to client for message", _messageCounter)
}
