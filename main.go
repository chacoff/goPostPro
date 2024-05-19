/*
 * File:    main.go
 * Date:    May 11, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Gathers data from thermal cameras at Train2 and cross-match with timestamps coming from MES to
 *	 to outcome post processes data.
 *
 * - Basic Build:
 *    go build -o ./Build/goPostPro.exe
 *
 * - Advance Build: require xmlstarlet: https://xmlstar.sourceforge.net/
 *    ./buildMachine.bat
 *     use later LaunchgoPostPro.lnk shortcut to start the application
 *
 * - Install as Windows Service:
 *    nssm install goPostProTr2
 *
 */

package main

import (
	"fmt"
	diasHelpers "goPostPro/dias"
	"goPostPro/global"
	mesHelpers "goPostPro/mes"
	"goPostPro/postpro"
	server "goPostPro/tcpServer"
	"log"
	"sync"
	"time"
)

// LTC default
var LTC []uint16 = []uint16{500, 501, 500, 502, 44, 55, 66, 77}

// init function starts Logger and DataBase
func init() {
	global.ConfigInit()
	fmt.Println(global.AppParams.Cage)

	log.Printf("***** Build Version: v %s %s *****", global.BuildParams.Version, global.BuildParams.Type)
	log.Printf("[livePostPro] init at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	errPostPro := postpro.StartDatabase()
	if errPostPro != nil {
		log.Panicln("error initializing DataBase")
	}
}

func main() {

	// Init servers
	dias := server.NewServer(global.AppParams.AddressDias, "DIAS")
	mes := server.NewServer(global.AppParams.Address, "MES")

	// Define waiting groups
	var wg sync.WaitGroup
	wg.Add(4)

	LTCch := make(chan []uint16) // LTC channel

	// Dias messages
	go func() {
		defer wg.Done()
		for msg := range dias.Msgch {
			// Dias payload
			_msg, _length := diasHelpers.DataScope(msg.Payload)
			diasHelpers.ProcessDiasData(msg.Payload)

			if global.AppParams.Verbose {
				log.Printf("[DIAS] received message length %d from (%s): %s\n", _length, msg.From, _msg)
			}

			// LTC consumer
			select {
			case ltc := <-LTCch:
				LTC = ltc
			default:
				//
			}

			_, err := msg.Conn.Write(diasHelpers.EncodeToDias(LTC))
			if err != nil {
				log.Printf("[DIAS] error writing response: %s\n", err)
				break
			}

			if global.AppParams.Verbose {
				log.Printf("[DIAS] sent to Dias %q\n", LTC)
			}
		}
	}()

	// MES messages
	go func() {
		defer wg.Done()
		for msg := range mes.Msgch {
			_payload, _len := mesHelpers.DataScope(msg.Payload)
			log.Printf("[MES] received message from %s with length %d: %s", msg.From, _len, _payload)

			header, hexBody := mesHelpers.HandleMesData(msg.Payload)
			echo, response, dataLTC, msgType, msgCounter := mesHelpers.HandleAnswerToMes(header, hexBody)

			// LTC producer
			switch msgType {
			case 4704, 4714: // process message: header + LTC - Cage3 and Cage4 only
				LTCch <- dataLTC // []uint16{500, 1200, 500, 1250, 44, 55, 66, 77}
			default:
				//
			}

			if echo {
				_, err := msg.Conn.Write(response)
				if err != nil {
					log.Println("[MES] error writing:", err)
					return
				}
				log.Println("[MES] response sent to client for message", msgCounter)
			}
		}
		close(LTCch)
	}()

	// Dias server start
	go func() {
		defer wg.Done()
		if err := dias.Start(); err != nil {
			log.Panicln("dias server error:", err)
		}
	}()

	// MES server start
	go func() {
		defer wg.Done()
		if err := mes.Start(); err != nil {
			log.Panicln("mes server error:", err)
		}
	}()

	wg.Wait()

}
