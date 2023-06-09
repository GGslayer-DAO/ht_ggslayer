// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameAdmin = "admins"

// Admin mapped from table <admins>
type Admin struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                      // 主键ID
	UserName  string    `gorm:"column:user_name;not null" json:"user_name"`                             // 用户名
	Email     string    `gorm:"column:email;not null" json:"email"`                                     // 邮箱
	Password  string    `gorm:"column:password" json:"password"`                                        // 密码
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"` // 更新时间
}

// TableName Admin's table name
func (*Admin) TableName() string {
	return TableNameAdmin
}

func NewAdmin() *Admin {
	return &Admin{}
}

func (m *Admin) Create() {
	db.Create(m)
}

//FindByUserName 根据用户名查找
func (m *Admin) FindByUserName(username string) (admins *Admin, err error) {
	err = db.Model(m).Where("user_name = ?", username).Find(&admins).Error
	return
}
