package dias

import (
	"encoding/binary"
	"encoding/hex"
	"goPostPro/global"
	"goPostPro/postpro"
	"log"
	"net"
)

var buffer = make([]byte, global.AppParams.MaxBufferSize)

// DiasServer opens socket communication with DIAS software. Objective is to pass the LTC value
func DiasServer() {

	ln, err := net.Listen(global.AppParams.NetType, global.AppParams.AddressDias)
	if err != nil {
		log.Printf("[DIAS SERVER] problems listening: %s\n", err)
	}
	defer ln.Close()
	log.Printf("[DIAS SERVER] listening DIAS on port: %s\n", global.AppParams.AddressDias)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[DIAS SERVER] error accepting connection: %s\n", err)
			continue
		}
		log.Printf("[DIAS SERVER] accepted DIAS-client")

		go handleDiasConnection(conn)

	}
}

func handleDiasConnection(conn net.Conn) {

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("[DIAS SERVER] error reading from connection: %s\n", err)
			break
		}

		// RECEIVED from DIAS
		message := buffer[:n]
		lenDias := len(buffer[:n])
		measurement_array := make([]int16, 0)

		for index := 2; index < len(message)-2; index += 2 {
			measurement_array = append(measurement_array, int16(binary.LittleEndian.Uint16(message[index:index+2])))
		}

		process_error := postpro.Process_live_line(measurement_array)
		if process_error != nil {
			log.Printf("[PROCESSING] error: %s\n", process_error)
		} else if global.AppParams.Verbose {
			log.Printf("[PROCESSING] completed: %d\n", measurement_array)
		}

		if global.AppParams.Verbose {
			log.Printf("[DIAS SERVER] len: %d received from Dias: %s\n", lenDias, message)
			log.Printf("[DIAS SERVER] new LTC values: %d\n", global.LTCFromMes)
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
			log.Printf("[DIAS SERVER] error writing response: %s\n", err)
			break
		} else {
			_length := len(answer)
			if global.AppParams.Verbose {
				log.Printf("[DIAS SERVER] sent to Dias %q with length: %d\n", hex.EncodeToString(answer), _length)
			}
		}
	}

	conn.Close()
}
