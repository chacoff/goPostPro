package dias

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
)

var buffer = make([]byte, 24)

// LTCServer opens socket communication with DIAS software. Objective is to pass the LTC value
func LTCServer(netType string, address string) {
	ln, err := net.Listen(netType, address)
	if err != nil {
		log.Fatal("problems listening: ", err)
	}
	defer ln.Close()
	fmt.Println("Listening DIAS on port: 127.0.0.1:2002")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err)
			continue
		}
		fmt.Println("Accepted DIAS-client")
		go handleDiasConnection(conn)
	}
}

func handleDiasConnection(conn net.Conn) {

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("error reading from connection: ", err)
			break
		}

		message := hex.EncodeToString(buffer[:n])
		fmt.Println("Message Received: ", message)

		answer := make([]byte, 0)
		values := []uint16{1501, 605, 706, 808, 609, 753, 855, 1165}
		for _, val := range values {
			binaryValue := make([]byte, 2)
			binary.LittleEndian.PutUint16(binaryValue, val)
			answer = append(answer, binaryValue...)
		}

		_, err = conn.Write(answer)
		if err != nil {
			fmt.Println("error writing response: ", err)
			break
		} else {
			_length := len(answer)
			fmt.Printf("Sent %q with length: %d\n", hex.EncodeToString(answer), _length)
		}
	}

	conn.Close()
}
