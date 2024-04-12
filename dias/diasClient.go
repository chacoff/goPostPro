package dias

import (
	"bufio"
	"fmt"
	"net"
)

//lint:ignore U1000 Ignore unused function temporarily for debugging
func diasClient() {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2002")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(conn, "From the client\n")

	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print(message)

	defer conn.Close()
}
