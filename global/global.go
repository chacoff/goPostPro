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
)

// Appconfig Main Software
var (
	Appconfig     config.Parameters = config.LoadConfig()
	AppParams     config.Config     = Appconfig.Config
	PostProParams config.PostPro    = Appconfig.PostPro
	DBParams      config.DataBase   = Appconfig.DataBase
	LogParams     config.LogParams  = Appconfig.Logs
)
