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
	"goPostPro/dias"
	"goPostPro/global"
	"goPostPro/mes"
	"goPostPro/postpro"
	"log"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	setConsole(global.Appconfig.Cage)

	loggerInit()
	postpro.Start_database()

	var wg sync.WaitGroup
	wg.Add(2)

	// dias-Server
	go func() {
		defer wg.Done()
		dias.DiasServer()
	}()

	// MES-Server
	go func() {
		defer wg.Done()
		mes.MESserver()
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

func loggerInit() {
	// rotation settings
	logger := &lumberjack.Logger{
		Filename:   "logs/livePostPro.log",
		MaxSize:    10,    // max. size in megas of the log file before it gets rotated
		MaxBackups: 5,    // max. number of old log files to keep
		MaxAge:     30,   // max. number of days to retain old log files
		Compress:   true, // compress the old log files
	}

	log.SetOutput(logger)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	log.Println("[LOGGER] Logs init.")
	defer logger.Close()
}