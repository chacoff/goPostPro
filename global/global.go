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
