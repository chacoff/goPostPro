package main

import (
	"encoding/hex"
	"fmt"
	"net"
)

// global variables
const (
	verbose       bool   = false
	netType       string = "tcp"
	address       string = "127.0.0.1:4600" // 10.28.114.89
	MaxBufferSize int    = 2048
	headerSize    int    = 40
)

var buffer []byte = make([]byte, MaxBufferSize) // buffer to hold incoming data
var allHexBytes []byte = make([]byte, MaxBufferSize)

func main() {
	listener, err := net.Listen(netType, address) // listen on port 4600
	if err != nil {
		fmt.Println("[WARNING] Error listening:", err)
		return
	}

	// defer listener.Close() // defer schedule functions to be called
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Printf("[WARNING] Error listening on port %s\n", err)
		}
	}(listener)

	fmt.Printf("Server is listening on port %s\n", address)
	for {
		conn, err := listener.Accept() // accept a connection
		if err != nil {
			fmt.Println("[WARNING] Error accepting connection:", err)
			break
		}

		go handleConnection(conn) // handle the connection in a goroutine
	}
}

func handleConnection(conn net.Conn) {

	// defer conn.Close() // close the connection when the function returns using a schedule: defer
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("[WARNING] error closing the connection %d\n", err)
		}
	}(conn)

	fmt.Println("Accepted connection from", conn.RemoteAddr())

	bytesRead := 0
	for bytesRead <= headerSize {
		n, err := conn.Read(buffer[bytesRead:]) // read data from the connection
		if err != nil {
			fmt.Println("[WARNING] Error reading or Client disconnected:", err)
			break
		}
		bytesRead += n
	}

	allHexBytes = buffer[:bytesRead]

	hexBytesHeader := allHexBytes[:headerSize]         // Extract first 40 bytes
	headerValues := decodeHeaderUint32(hexBytesHeader) // decode little-endian uint16 values
	fmt.Println(">> Decoded Header values:", headerValues)

	hexBytesBody := allHexBytes[headerSize:] // Extract the rest of the bytes

	if verbose {
		hexData := hex.EncodeToString(buffer[:bytesRead]) // convert binary data to hexadecimal representation
		hexBytes, err := hex.DecodeString(hexData)        // back to bytes
		if err != nil {
			fmt.Println("[WARNING] Error decoding hex string:", err)
			return
		}

		fmt.Printf(">> Received data (hex): %s\n", hexData)      // print the received data in hexadecimal format
		fmt.Printf("- Number of characters: %d\n", len(hexData)) // count characters
		fmt.Printf("- Number of bytes: %d\n", len(hexBytes))     // count bytes
	}

	go handleAnswer(conn, headerValues, hexBytesBody)
}

func handleAnswer(conn net.Conn, _headerValues []uint32, _hexBytesBody []byte) {

	var echo bool = false
	var response []byte
	messageType := int(_headerValues[1]) // message type on the header
	messageCounter := _headerValues[2]   // already in uint32
	messageTypeAns := uint32(messageType - 100)

	switch messageType {
	case 4701: // watchdog, not a body to decode
		response = encodeUint32(headerType(40, messageTypeAns, messageCounter))
		echo = true

	case 4702: // process message
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

	if echo {
		_, err := conn.Write(response)
		if err != nil {
			fmt.Println("Error writing:", err)
			return
		}
		fmt.Println("Response sent to client for message", messageCounter)
	}
}
