package dias

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"goPostPro/global"
	"log"
	"net"
	// "time"
)

var buffer = make([]byte, global.Appconfig.MaxBufferSize)

// DiasServer opens socket communication with DIAS software. Objective is to pass the LTC value
func DiasServer() {

	ln, err := net.Listen(global.Appconfig.NetType, global.Appconfig.AddressDias)
	if err != nil {
		log.Fatal("[DIAS SERVER] problems listening: ", err)
	}
	defer ln.Close()
	fmt.Printf("[DIAS SERVER] listening DIAS on port: %s\n", global.Appconfig.AddressDias)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[DIAS SERVER] error accepting connection: ", err)
			continue
		}
		fmt.Println("[DIAS SERVER] accepted DIAS-client")

		go handleDiasConnection(conn)

	}
}

func handleDiasConnection(conn net.Conn) {

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("[DIAS SERVER] error reading from connection: ", err)
			break
		}

		// RECEIVED from DIAS
		message := hex.EncodeToString(buffer[:n])
		lenDias := len(buffer[:n])
		if global.Appconfig.Verbose {
			fmt.Printf("[DIAS SERVER] len: %d received from Dias: %s ", lenDias, message)
			fmt.Println("[DIAS SERVER] new LTC values: ", global.LTCFromMes)
		}

		// SENT to DIAS
		LTCValues := global.LTCFromMes

		answer := make([]byte, 0)
		for _, val := range LTCValues {
			binaryValue := make([]byte, 2)
			binary.LittleEndian.PutUint16(binaryValue, val)
			answer = append(answer, binaryValue...)
		}

		_, err = conn.Write(answer)
		// time.Sleep(1000 * time.Millisecond)
		if err != nil {
			fmt.Println("[DIAS SERVER] error writing response: ", err)
			break
		} else {
			_length := len(answer)
			if global.Appconfig.Verbose{
				fmt.Printf("[DIAS SERVER] sent to Dias %q with length: %d\n", hex.EncodeToString(answer), _length)
			}
			
		}
	}

	conn.Close()
}
