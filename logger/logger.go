package logger

import (
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger zerolog.Logger
)


func init() {
	// Set up log rotation settings
	logRotation := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    1,    // max. size in megas of the log file before it gets rotated
		MaxBackups: 5,    // max. number of old log files to keep
		MaxAge:     30,   // max. number of days to retain old log files
		Compress:   true, // compress the old log files
	}

	logger = zerolog.New(logRotation).With().Timestamp().Caller().Logger()
	zerolog.TimeFieldFormat = time.RFC3339 // timeFormat

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		_, filename := filepath.Split(file)
		return filename
	}

	logger.Info().Msg("Log init ...")
	// logger.Error().Msg("This is an error message")
}

func Warning(text string){
	logger.Warn().Msg(text)
}