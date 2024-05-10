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
