package dias

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"goPostPro/global"
	"log"
	"net"
	"time"
)

var buffer = make([]byte, 24)

// DiasServer opens socket communication with DIAS software. Objective is to pass the LTC value
func DiasServer(valuesToDias <-chan []uint16) {

	var LTCValues []uint16

	ln, err := net.Listen(global.Appconfig.NetType, global.Appconfig.AddressDias)
	if err != nil {
		log.Fatal("problems listening: ", err)
	}
	defer ln.Close()
	fmt.Printf("Listening DIAS on port: %s\n", global.Appconfig.AddressDias)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err)
			continue
		}
		fmt.Println("Accepted DIAS-client")

		go handleDiasConnection(conn, LTCValues)

		select {
		case LTCValues = <-valuesToDias:
			fmt.Println("DIAS: Received data from channel:", LTCValues) // Data is available on the channel, process it
		default:
			LTCValues = []uint16{1,2,3,4,5,6,7,8} // dummy LTCs if the channel is empty
			fmt.Println("DIAS: No data available on the channel. Using defaults: ", LTCValues)  // No data available on the channel, do something else
		}
	}
}

func handleDiasConnection(conn net.Conn, LTCValues []uint16) {

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("error reading from connection: ", err)
			break
		}

		message := hex.EncodeToString(buffer[:n])
		if global.Appconfig.Verbose {
			fmt.Println("Message Received from Dias: ", message)
		}
		fmt.Println("Updating Dias:", LTCValues)

		answer := make([]byte, 0)
		for _, val := range LTCValues {
			binaryValue := make([]byte, 2)
			binary.LittleEndian.PutUint16(binaryValue, val)
			answer = append(answer, binaryValue...)
		}

		_, err = conn.Write(answer)
		time.Sleep(1200 * time.Millisecond)
		if err != nil {
			fmt.Println("error writing response: ", err)
			break
		} else {
			_length := len(answer)
			fmt.Printf("Sent to Dias %q with length: %d\n", hex.EncodeToString(answer), _length)
		}
	}

	conn.Close()
}
