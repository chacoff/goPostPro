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
	"goPostPro/config"
	"goPostPro/dias"
	"goPostPro/mes"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var appconfig config.Config

func main() {
	appconfig = config.LoadConfig()
	setConsoleTitle(appconfig.Cage)

	var wg sync.WaitGroup
	wg.Add(2)

	// dias-Server
	go func(){
		defer wg.Done()
		go dias.LTCServer(appconfig.NetType, appconfig.AddressDias)
	}()
	

	// MES-Server
	go func(){
		defer wg.Done()
		go mes.MESserver()
	}()
	
	// PLC-client
	// go plc.SiemensClient()

	wg.Wait()

}

func setConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTitleW")

	titleUTF16 := utf16.Encode([]rune(title + "\x00"))

	proc.Call(uintptr(unsafe.Pointer(&titleUTF16[0])))
}
