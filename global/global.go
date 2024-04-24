package global

import (
	"goPostPro/config"
)

var Appconfig config.Config = config.LoadConfig()

var LTCFromMes []uint16 // TODO only a workaround!
