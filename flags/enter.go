package flags

import (
	"flag"
)

type Option struct {
	DB   bool
	Api  bool
	Port int
}

// Parse 解析命令行参数
func Parse() (option *Option) {
	option = new(Option)
	flag.BoolVar(&option.DB, "db", false, "初始化数据库")
	flag.BoolVar(&option.Api, "api", false, "Start a api server?")
	flag.IntVar(&option.Port, "port", 10001, "port")
	flag.Parse()
	return option
}

// Run 根据命令执行不同的函数
func (option Option) Run() bool {
	if option.DB {
		DB()
		return true
	}
	return false
}
