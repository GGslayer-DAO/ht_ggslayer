// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"ggslayer/utils"
	"time"
)

const TableNameUser = "users"

// User mapped from table <users>
type User struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"-"`                // 主键ID
	Name      string    `gorm:"column:name;not null" json:"name"`                                // 用户名
	Email     string    `gorm:"column:email;not null" json:"email"`                              // 邮箱
	GoogleEmail     string    `gorm:"column:google_email;null" json:"-"`                              // 邮箱
	HeadPic   string    `gorm:"column:headpic;null" json:"headpic"`                              // 头像
	Cn        string    `gorm:"column:cn" json:"cn"`                                             // 国号
	Phone     string    `gorm:"column:phone" json:"phone"`                                       // 电话号码
	Describe  string    `gorm:"column:describe" json:"describe"`                                 // 简介
	Age       int32     `gorm:"column:age;not null" json:"age"`                                  // 年龄
	Distinct  string    `gorm:"column:distinct" json:"distinct"`                                 // 地区
	Point     int32     `gorm:"column:point;not null" json:"point"`                              // 积分
	Exp       int32     `gorm:"column:exp;not null" json:"exp"`                                  // 经验值
	Password  string    `gorm:"column:password" json:"-"`                                        // 密码
	Status    int32     `gorm:"column:status;not null" json:"-"`                                 // 状态
	Level     int32     `gorm:"column:level;not null" json:"level"`                              // 等级
	ReferCode string     `gorm:"column:refer_code;not null" json:"refer_code"`                    // 邀请码
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"-"`   // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:1970-01-01 08:00:01" json:"-"` // 更新时间
	UserToken []*UserToken `gorm:"-" json:"user_token"`
	Follower    int    `gorm:"-" json:"follower"`
	Following  int   `gorm:"-" json:"following"`
	Likes  int   `gorm:"-" json:"likes"`
}

func NewUser() *User {
	return &User{}
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}

//判断email是否存在
func (m *User) FindInfoByEmail(email string) (user *User, err error) {
	err = db.Model(m).Where("email = ?", email).Find(&user).Error
	return
}

//判断googleEmail是否存在
func (m *User) FindInfoByGoogleEmail(googleEmail string) (user *User, err error) {
	err = db.Model(m).Where("google_email = ?", googleEmail).Find(&user).Error
	return
}

//查找邀请码是否存在
func (m *User) FindReferCode(referCode string) (user *User) {
	db.Model(m).Where("refer_code = ?", referCode).Find(&user)
	return
}

//创建用户
func (m *User) Create() (userId int32, err error) {
	err = db.Create(m).Error
	return m.ID, err
}

//编辑更新用户
func (m *User) Update() error {
	return db.Updates(m).Error
}

//根据id查找用户数据信息
func (m *User) FindById(id int) (user *User, err error) {
	err = db.Model(m).Where("id = ?", id).Find(&user).Error
	return
}

//修改用户经验值和等级
func (m *User) ChangeUserExps(userId, exp int) (err error) {
	user, err := m.FindById(userId)
	if err != nil {
		return
	}
	userExp := utils.Add(float64(user.Exp), float64(exp), 0)
	if userExp >= 10 && userExp < 210 {
		user.Level = 1
	} else if userExp >= 210 && userExp < 1380 {
		user.Level = 2
	} else if userExp >= 1380 && userExp < 3630 {
		user.Level = 3
	} else if userExp >= 3630 && userExp < 6960 {
		user.Level = 4
	} else if userExp >= 6960 && userExp < 11370 {
		user.Level = 5
	} else if userExp >= 11370 && userExp < 16860 {
		user.Level = 6
	} else if userExp >= 16860 && userExp < 23430 {
		user.Level = 7
	} else if userExp >= 23430 && userExp < 31080 {
		user.Level = 8
	} else if userExp >= 31080 {
		user.Level = 9
	}
	user.Exp = int32(userExp)
	err = user.Update()
	return
}

//FindHeadPicByUserId 根据用户id查找用户头像
func (m *User) FindHeadPicByUserId(userId []int) (headpics []string, err error) {
	err = db.Model(m).Where("id in (?)", userId).Pluck("headpic", &headpics).Error
	return
}

//FindUserMapByUserIds 根据用户id数组查找用户信息
func (m *User) FindUserMapByUserIds(userId []int32) (userMap map[int32]*User, err error) {
	var users []*User
	err = db.Model(m).Where("id in (?)", userId).Find(&users).Error
	if err != nil {
		return
	}
	userMap = make(map[int32]*User)
	for _, user := range users {
		userMap[user.ID] = user
	}
	return
}