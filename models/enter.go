package models

import "time"

type MODEL struct {
	ID        uint      `gorm:"primaryKey;comment:id" json:"id"` // 主键ID
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`  // 创建时间
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`  // 更新时间
}
