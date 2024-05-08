package global

import (
	"goPostPro/config"
)

// Appconfig Main Software
var (
	Appconfig     config.Parameters = config.LoadConfig()
	AppParams     config.Config     = Appconfig.Config
	PostProParams config.PostPro    = Appconfig.PostPro
	DBParams      config.DataBase   = Appconfig.DataBase
	LogParams     config.LogParams  = Appconfig.Logs
)

var LTCFromMes []uint16 = []uint16{500, 950, 500, 980, 44, 55, 66, 77} // TODO only a workaround!

//// Post Processing
//var (
//	TIME_FORMAT                   string  = "2006-01-02 15:04:05,999" // Format the time as ISO 8601
//	NUMBER_FIRST_MEASURES_REMOVED int     = 5
//	TEMPERATURE_THRESHOLD_FACTOR  float64 = 0.35
//	TEMPERATURE_THRESHOLD_MINIMUM float64 = 780
//	GRADIENT_LIMIT_FACTOR         float64 = 3
//	WIDTH_MINIMUM                 int64   = 2
//)
//
//var (
//	DATABASE_PATH        string = "./processed.db"
//	TIME_FORMAT_REQUESTS string = "20060102150405" // REFERENCE FOR TIMESTAMP!
//)
