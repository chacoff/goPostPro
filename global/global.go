package global

import (
	"goPostPro/config"
)

var Appconfig config.Config = config.LoadConfig()

var LTCFromMes []uint16 = []uint16{1, 2, 3, 4, 5, 6, 7, 8} // TODO only a workaround!
