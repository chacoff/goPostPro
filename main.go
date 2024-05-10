/*
 * File:    main.go
 * Date:    March 04, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Gathers data from thermal cameras at Train2 and cross-match with timestamps coming from MES to
 *	 to outcome post processes data.
 *
 * Build:
 * go build -o ./Build/goPostPro.exe
 */

package main

import (
	diasHelpers "goPostPro/dias"
	"goPostPro/global"
	mesHelpers "goPostPro/mes"
	"goPostPro/postpro"
	server "goPostPro/tcpServer"
	"log"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LTC default
var LTC []uint16 = []uint16{500, 501, 500, 502, 44, 55, 66, 77}

func main() {
	// Init
	setConsole(global.AppParams.Cage)

	errLogger := loggerInit()
	if errLogger != nil {
		log.Panicln("error initializing Logger")
	}

	errPostPro := postpro.StartDatabase()
	if errPostPro != nil {
		log.Panicln("error initializing DataBase")
	}

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

func setConsole(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTitleW")

	titleUTF16 := utf16.Encode([]rune(title + "\x00"))

	_, _, err := proc.Call(uintptr(unsafe.Pointer(&titleUTF16[0])))
	if err != nil {
		return
	}
}

func loggerInit() error {
	// rotation settings
	logger := &lumberjack.Logger{
		Filename:   global.LogParams.FileName,
		MaxSize:    global.LogParams.MaxSize,    // max. size in megas of the log file before it gets rotated
		MaxBackups: global.LogParams.MaxBackups, // max. number of old log files to keep
		MaxAge:     global.LogParams.MaxAge,     // max. number of days to retain old log files
		Compress:   global.LogParams.Compress,   // compress the old log files
	}

	log.SetOutput(logger)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	log.Println("** ----------------------------------------------------------------")
	log.Printf("[livePostPro] init at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	defer logger.Close()

	return nil
}
