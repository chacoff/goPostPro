package dias

import (
	"bufio"
	"fmt"
	"net"
)

func DiasClient() {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:5603")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(conn, "From the client\n")

	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print(message)

	defer conn.Close()
}
