package core

import (
	"fmt"
	"ggcache-plus/config"
	"ggcache-plus/global"
	"github.com/spf13/viper"
	"log"
)

const ConfigFile = "settings"

func init() {
	viper.SetConfigName(ConfigFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("get yamlConf error: %s", err))
	}
}

// InitConf 读取yaml文件的配置
func InitConf() {
	c := &config.Config{}
	err := viper.Unmarshal(c)
	if err != nil {
		log.Fatal("config Init Unmarshal: %v", err)
	}
	log.Println("配置文件加载成功")
	global.Config = c
}

// SetYaml 修改yaml文件
func SetYaml(key string, value any) error {
	viper.Set(key, value)
	err := viper.WriteConfig()
	if err != nil {
		global.Log.Error("配置文件修改失败")
		return err
	}
	global.Log.Info("配置文件修改成功")
	return nil
}
