package main

import (
	"fmt"
	"ggcache-plus/core"
	"ggcache-plus/global"
)

func main() {
	core.InitConf()
	core.InitEtcd()
	global.Log = core.InitLogger()

	fmt.Println(global.Config.GGroupCache)
	fmt.Println(global.DefaultEtcdConfig)
}
