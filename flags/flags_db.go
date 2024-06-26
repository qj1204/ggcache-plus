package flags

import (
	"ggcache-plus/global"
	"ggcache-plus/models"
)

func DB() {
	var err error
	// 生成四张表的表结构
	err = global.DB.Set("gorm:table_options", "ENGINE=InnoDB").
		AutoMigrate(
			&models.StudentModel{},
		)
	if err != nil {
		global.Log.Error("[error] 生成数据库表结构失败")
		return
	}
	global.Log.Info("[success] 生成数据库表结构成功！")
}
