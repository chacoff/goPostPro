package dias

import (
	"bufio"
	"fmt"
	"net"
)

<<<<<<< HEAD
// diasClient client for testing DIAS server. it is now deprecated, we use DIAS tester
func diasClient() {
=======
func DiasClient() {
>>>>>>> 6a939b613138968b477dec68b92cadb95dfb2d2d
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
