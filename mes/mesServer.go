/*
 * File:    mesServer.go
 * Date:    April 17, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Open a tcp ip server communication with MES at train2
 *
 */

package mes

import (
	"fmt"
	"net"
	"os"
	"sync"

	"goPostPro/global"
)

var (
	buffer      = make([]byte, 2048)    // buffer to hold incoming data
	allHexBytes = make([]byte, 0, 2048) // variable: from 0 to maxBufferSize. Data will be appended on arrival
)

// MESserver to receive the messages from MES
func MESserver(valuesToDias chan<- []uint16) {

	var resultDias []uint16
	var wg1 sync.WaitGroup

	listener, err := net.Listen(global.Appconfig.NetType, global.Appconfig.Address) // listen on port 4600
	if err != nil {
		fmt.Println("[WARNING] Error listening:", err)
		// return
		os.Exit(1)
	}
	defer listener.Close() // close the connection when the function returns using a schedule: defer
	fmt.Printf("Listening MES on port: %s\n", global.Appconfig.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[WARNING] Error accepting connection:", err)
			os.Exit(1)
			// break
		}

		fmt.Println("Accepted MES-client from", conn.RemoteAddr())
		wg1.Add(1)
		go handleConnection(conn, &resultDias, &wg1)
		wg1.Wait()
		valuesToDias <- resultDias
	}
}

func handleConnection(conn net.Conn, resultDias *[]uint16, wg *sync.WaitGroup) {

	defer conn.Close()
	defer wg.Done()

	var isHeaderOk = false
	var headerValues []uint32
	var FullLength int

	var tempResultDias []uint16
	var wg2 sync.WaitGroup

	for {
		n, err := conn.Read(buffer) // read data from the connection
		if err != nil {             // && err != io.EOF
			fmt.Println("[WARNING] Error reading or Client disconnected:", err)
			allHexBytes = make([]byte, 0, global.Appconfig.MaxBufferSize) // reset variable before handling answer
			isHeaderOk = false                                            // reset variables before handling answer
			break
		}
		allHexBytes = append(allHexBytes, buffer[:n]...) //  append data upon arrival

		if len(allHexBytes) >= global.Appconfig.HeaderSize && !isHeaderOk {
			hexBytesHeader := allHexBytes[:global.Appconfig.HeaderSize]   // Extract first 40 bytes, only header
			headerValues, isHeaderOk = decodeHeaderUint32(hexBytesHeader) // decode little-endian uint32 values
			FullLength = int(headerValues[0])
			fmt.Println(">> Decoded Header values:", headerValues)
		}

		if len(allHexBytes) >= FullLength && isHeaderOk { // TODO attention with the '>='
			hexBytesBody := allHexBytes[global.Appconfig.HeaderSize:FullLength] // Extract the rest of the bytes
			allHexBytes = make([]byte, 0, global.Appconfig.MaxBufferSize)       // reset variable before handling answer
			isHeaderOk = false                                                  // reset variables before handling answer
			wg2.Add(1)
			go handleAnswer(conn, headerValues, hexBytesBody, &tempResultDias, &wg2)
			wg2.Wait()
			*resultDias = tempResultDias
		}
	}
}

func handleAnswer(conn net.Conn, _headerValues []uint32, _hexBytesBody []byte, resultLTC *[]uint16, wg *sync.WaitGroup) {

	defer wg.Done()

	var echo = false
	var response []byte
	var dataLTC []uint16 = []uint16{0, 1, 2, 3, 4, 5, 6, 7} // default values

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
		//bodyValuesStatic, _ := decodeBody(_hexBytesBody, messageType)
		//fmt.Println(">> Decoded LTC values:", bodyValuesStatic)
		fmt.Println("LTC received")
		dataLTC = []uint16{uint16(_headerValues[3]), 1234, 5678, 7891, 7895, 750, 850, uint16(_headerValues[4])}

		echo = false

	case 4703, 4713, 4723: // acknowledge data message
		fmt.Println("MES received data properly")
		echo = false

	default:
		fmt.Println("Unknown message:", messageType, messageCounter)
		echo = false
	}

	*resultLTC = dataLTC // LTC data DIAS coming from MES

	if echo {
		_, err := conn.Write(response)
		if err != nil {
			fmt.Println("Error writing:", err)
			return
		}
		fmt.Println("Response sent to client for message", messageCounter)
	}
}
