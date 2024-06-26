package main

import (
	"ggcache-plus/core"
	"ggcache-plus/global"
)

func main() {
	core.InitConf()
	global.Log = core.InitLogger()

	core.SetYaml("mysql.port", 3307)
}
