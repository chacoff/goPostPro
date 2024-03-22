package dias

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func diasServer() {
	ln, err := net.Listen("tcp", "127.0.0.1:5603")
	if err != nil {
		fmt.Println("problems listening")
	}
	fmt.Println("Listen on port: 127.0.0.1:5603")

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

	var newMessage string

	for {
		message, er := bufio.NewReader(conn).ReadString('\n')
		if er != nil {
			fmt.Println("disconnection error")
			break
		}

		fmt.Println("Message Received:", message)
		newMessage = strings.ToUpper("from the Server")
	}

	conn.Write([]byte(newMessage))

}
