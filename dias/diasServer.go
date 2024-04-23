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

		data := <-valuesToDias
		go handleDiasConnection(conn, data)
	}
}

func handleDiasConnection(conn net.Conn, values []uint16) {

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
		fmt.Println("Updating Dias:", values)

		answer := make([]byte, 0)
		for _, val := range values {
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
