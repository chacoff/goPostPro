package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
)

var (
	i      int = 0
	buffer     = make([]byte, 2048)
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:2002")
	if err != nil {
		fmt.Println("problems listening")
	}
	fmt.Println("Listen on port: 127.0.0.1:2002")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Accepted connection on port")
		go handleDiasConnection(conn)
	}
}

func handleDiasConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("disconnection ...")
		}
	}(conn)

	_buffer := make([]byte, 0, 2048)
	for {
		fmt.Println(":Debug", i)
		n, err := conn.Read(buffer) // read data from the connection
		if err != nil {
			fmt.Println("error reading or disconnection")
			break
		}

		message := hex.EncodeToString(buffer[:n])
		fmt.Println("Message Received: ", message, len(buffer[:n]))

		values := []uint16{1500, 600, 700, 800, 600, 750, 850, 1160}
		for _, val := range values {
			binaryValue := make([]byte, 2)
			binary.LittleEndian.PutUint16(binaryValue, val)
			_buffer = append(_buffer, binaryValue...) // slice, element to unpack
		}
		i += 1

	}
	go handleAnswer(conn, _buffer)
}

func handleAnswer(conn net.Conn, _buffer []byte) {
	fmt.Println("length answer: ", len(_buffer))
	fmt.Println(hex.EncodeToString(_buffer))
	conn.Write(_buffer)
	fmt.Println("Sent.")
}
