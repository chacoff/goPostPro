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
	"goPostPro/dias"
	"net"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	config      Config
	buffer      = make([]byte, 2048)    // buffer to hold incoming data
	allHexBytes = make([]byte, 0, 2048) // variable: from 0 to maxBufferSize. Data will be appended on arrival
)

func main() {
	config = loadConfig()
	setConsoleTitle(config.Cage)

	// dias-Server
	go dias.LTCServer(config.NetType, config.AddressDias)

	// MES-Server
	listener, err := net.Listen(config.NetType, config.Address) // listen on port 4600
	if err != nil {
		fmt.Println("[WARNING] Error listening:", err)
		// return
		os.Exit(1)
	}
	defer listener.Close() // close the connection when the function returns using a schedule: defer
	fmt.Printf("Listening MES on port %s\n", config.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[WARNING] Error accepting connection:", err)
			os.Exit(1)
			// break
		}

		fmt.Println("Accepted MES-client from", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	var isHeaderOk = false
	var headerValues []uint32
	var FullLength int

	for {
		n, err := conn.Read(buffer) // read data from the connection
		if err != nil {             // && err != io.EOF
			fmt.Println("[WARNING] Error reading or Client disconnected:", err)
			allHexBytes = make([]byte, 0, config.MaxBufferSize) // reset variable before handling answer
			isHeaderOk = false                                  // reset variables before handling answer
			break
		}
		allHexBytes = append(allHexBytes, buffer[:n]...) //  append data upon arrival

		if len(allHexBytes) >= config.HeaderSize && isHeaderOk == false {
			hexBytesHeader := allHexBytes[:config.HeaderSize]             // Extract first 40 bytes, only header
			headerValues, isHeaderOk = decodeHeaderUint32(hexBytesHeader) // decode little-endian uint32 values
			FullLength = int(headerValues[0])
			fmt.Println(">> Decoded Header values:", headerValues)
		}

		if len(allHexBytes) >= FullLength && isHeaderOk == true { // TODO attention with the '>='
			hexBytesBody := allHexBytes[config.HeaderSize:FullLength] // Extract the rest of the bytes
			allHexBytes = make([]byte, 0, config.MaxBufferSize)       // reset variable before handling answer
			isHeaderOk = false                                        // reset variables before handling answer
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
	case 4701, 4711, 4721: // watchdog: only header
		response = encodeUint32(headerType(40, messageTypeAns, messageCounter))
		echo = true

	case 4702, 4712, 4722: // process message: header + body
		bodyValuesStatic, bodyValueDynamic := decodeBody(_hexBytesBody, messageType)
		fmt.Println(">> Decoded Body values:", bodyValuesStatic, bodyValueDynamic)

		_bodyAns := encodeProcess(processType(bodyValuesStatic, bodyValueDynamic))
		_length := uint32(40 + len(_bodyAns))
		_headerAns := encodeUint32(headerType(_length, messageTypeAns, messageCounter))

		_response := make([]byte, 0, len(_headerAns)+len(_bodyAns))
		_response = append(_response, _headerAns...)
		_response = append(_response, _bodyAns...)

		response = _response

		echo = true

	case 4704, 4714: // process message: header + LTC - Cage3 and Cage4 only
		bodyValuesStatic, _ := decodeBody(_hexBytesBody, messageType)
		fmt.Println(">> Decoded LTC values:", bodyValuesStatic)

		echo = false

	case 4703, 4713, 4723: // acknowledge data message
		fmt.Println("MES received data properly")
		echo = false

	default:
		fmt.Println("Unknown message:", messageType, messageCounter)
		echo = false
	}

	if echo == true {
		_, err := conn.Write(response)
		if err != nil {
			fmt.Println("Error writing:", err)
			return
		}
		fmt.Println("Response sent to client for message", messageCounter)
	}
}

func setConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTitleW")

	titleUTF16 := utf16.Encode([]rune(title + "\x00"))

	proc.Call(uintptr(unsafe.Pointer(&titleUTF16[0])))
}
