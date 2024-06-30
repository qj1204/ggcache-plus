package distributekv

import (
	"ggcache-plus/global"
	"ggcache-plus/models"
	"gorm.io/gorm"
	"strconv"
)

func NewGroupInstance(groupname string) *Group {
	g := NewGroup(groupname, "lru", 2<<10, RetrieverFunc(func(key string) ([]byte, error) {
		// 从后端数据库中查找
		global.Log.Info("进入 RetrieverFunc，在数据库中查询....")

		var students []*models.StudentModel
		count := global.DB.Where("name = ?", key).Find(&students).RowsAffected
		if count == 0 {
			global.Log.Info("数据库中没有该条记录...")
			return []byte{}, gorm.ErrRecordNotFound
		}

		global.Log.Infof("成功从数据库中查询到学生 %s 的分数为：%d", key, students[0].Score)
		return []byte(strconv.Itoa(students[0].Score)), nil
	}))
	groups[groupname] = g
	InitDataWithGroup(g)
	return g
}

func InitDataWithGroup(g *Group) {
	// 先往数据库中存一些数据（慢速数据库）
	global.DB.Create(&models.StudentModel{Name: "张三", Score: 333})
	global.DB.Create(&models.StudentModel{Name: "李四", Score: 444})
	global.DB.Create(&models.StudentModel{Name: "王五", Score: 555})
	global.Log.Info("---db 数据添加成功---")
	// 往缓存中存储一些元素
	g.mainCache.add("abc", ByteView{b: []byte("123")})
	g.mainCache.add("bcd", ByteView{b: []byte("456")})
	g.mainCache.add("cde", ByteView{b: []byte("789")})
	global.Log.Info("---cache 数据添加成功---")
}
