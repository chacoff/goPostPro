package main

import (
	"encoding/hex"
	"fmt"
	"net"
)

// global variables
const (
	verbose    bool   = false
	netType    string = "tcp"
	address    string = "127.0.0.1:4600"
	bufferSize int    = 2048
	headerSize int    = 40
)

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

	buffer := make([]byte, bufferSize) // buffer to hold incoming data

	for {
		n, err := conn.Read(buffer) // read data from the connection
		if err != nil {
			fmt.Println("[WARNING] Error reading or Client disconnected:", err)
			break
		}

		hexData := hex.EncodeToString(buffer[:n]) // convert binary data to hexadecimal representation

		hexBytes, err := hex.DecodeString(hexData)
		if err != nil {
			fmt.Println("[WARNING] Error decoding hex string:", err)
			return
		}

		hexBytesHeader := hexBytes[:headerSize] // Extract first 40 bytes
		// hexBytesBody := hexBytes[headerSize:]    // Extract the rest of the bytes

		if verbose {
			fmt.Printf(">> Received data (hex): %s\n", hexData) // print the received data in hexadecimal format
			fmt.Printf(">> Received header (hex): %s\n", hex.EncodeToString(hexBytesHeader))
			charactersCount := len(hexData) // count characters
			fmt.Printf("- Number of characters: %d\n", charactersCount)
			bytesCount := len(hexBytes) // count bytes
			fmt.Printf("- Number of bytes: %d\n", bytesCount)
		}

		values := decodeHeaderUint32(hexBytesHeader) // decode little-endian uint16 values
		fmt.Println(">> Decoded Header values:", values)
	}
}
