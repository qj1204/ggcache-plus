package models

type StudentModel struct {
	MODEL
	Name  string `gorm:"size:20;comment:学生姓名" json:"name"`
	Score int    `gorm:"comment:分数" json:"score"`
}
