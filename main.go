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
	"time"
	"unicode/utf16"
	"unsafe"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	setConsole(global.AppParams.Cage)

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
		Filename:   global.LogParams.FileName,
		MaxSize:    global.LogParams.MaxSize,    // max. size in megas of the log file before it gets rotated
		MaxBackups: global.LogParams.MaxBackups, // max. number of old log files to keep
		MaxAge:     global.LogParams.MaxAge,     // max. number of days to retain old log files
		Compress:   global.LogParams.Compress,   // compress the old log files
	}

	log.SetOutput(logger)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	log.Printf("---------------- [livePostPro] init at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	defer logger.Close()
}
