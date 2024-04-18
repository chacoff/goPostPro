/*
 * File:    main.go
 * Date:    March 04, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Gathers data from thermal cameras at Train2 and cross-match with timestamps coming from MES to
 *	 to outcome post processes data.
 */

package main

import (
	"goPostPro/dias"
	"goPostPro/global"
	"goPostPro/mes"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

func main() {
	setConsoleTitle(global.Appconfig.Cage)

	var wg sync.WaitGroup
	wg.Add(2)

	valuesToDias := make(chan []uint16)

	// dias-Server
	go func() {
		defer wg.Done()
		dias.DiasServer(valuesToDias)
	}()

	// MES-Server
	go func() {
		defer wg.Done()
		mes.MESserver(valuesToDias)
	}()

	// PLC-client
	// go plc.SiemensClient()

	wg.Wait()

	close(valuesToDias)

}

func setConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTitleW")

	titleUTF16 := utf16.Encode([]rune(title + "\x00"))

	proc.Call(uintptr(unsafe.Pointer(&titleUTF16[0])))
}
