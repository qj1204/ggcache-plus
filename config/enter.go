package config

type Config struct {
	Mysql         Mysql
	LoggerSetting LoggerSetting `mapstructure:"logger_setting"`
	Etcd          Etcd
	GGroupCache   GGroupCache
}
