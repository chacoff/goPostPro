/*
 * File:    global.go
 * Date:    May 11, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   a package with global variables for the whole project
 *
 */

package global

import (
	"goPostPro/config"
	"log"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Appconfig Main Software
var (
	AppParams     	config.Config
	PostProParams 	config.PostPro
	DBParams      	config.DataBase
	LogParams     	config.LogParams
	IsConsoleApp	bool = true
)

// ConfigInit public method that initialize the config variables and the logger
func ConfigInit(){
	Appconfig, err := config.LoadConfig()
	if err != nil{
		log.Fatalf("Error loading configurations: %s\n", err)
	}

	AppParams = Appconfig.Config
	PostProParams = Appconfig.PostPro
	DBParams = Appconfig.DataBase
	LogParams = Appconfig.Logs

	errLogger := loggerInit()
	if errLogger != nil {
		log.Panicln("error initializing Logger")
	}

}

// loggerInit private method loading the parameters for the logger
func loggerInit() error {
	// rotation settings
	logger := &lumberjack.Logger{
		Filename:   LogParams.FileName,
		MaxSize:    LogParams.MaxSize,    // max. size in megas of the log file before it gets rotated
		MaxBackups: LogParams.MaxBackups, // max. number of old log files to keep
		MaxAge:     LogParams.MaxAge,     // max. number of days to retain old log files
		Compress:   LogParams.Compress,   // compress the old log files
	}

	log.SetOutput(logger)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	defer logger.Close()

	return nil
}

// SetConsole is a method to define a name when the software is used as a consoleApp
func SetConsole(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTitleW")

	titleUTF16 := utf16.Encode([]rune(title + "\x00"))

	_, _, err := proc.Call(uintptr(unsafe.Pointer(&titleUTF16[0])))
	if err != nil {
		return
	}
}

