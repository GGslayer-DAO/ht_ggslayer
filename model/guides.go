// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"fmt"
	"time"
)

const TableNameGuide = "guides"

// Guide mapped from table <guides>
type Guide struct {
	ID          int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                      // 主键ID
	GameID      int32     `gorm:"column:game_id;not null" json:"game_id"`                                 // 游戏id
	Title       string    `gorm:"column:title;not null" json:"title"`                                     // 标题
	Describe    string    `gorm:"column:describe" json:"describe"`                                        // 简介
	Content     string    `gorm:"column:content" json:"content"`                                          // 内容
	Rate        float64   `gorm:"column:rate;not null;default:0.00" json:"rate"`                          // 攻略评分
	URL         string    `gorm:"column:url;not null" json:"url"`                                         // 图片链接封面
	Author      string    `gorm:"column:author" json:"author"`                                            // 作者
	AuthorImage string    `gorm:"column:author_image" json:"author_image"`                                // 作者头像
	Visit       int     `gorm:"column:visit;not null" json:"visit"`                                     // 浏览数量
	Collect     int     `gorm:"column:collect;not null" json:"collect"`                                 // 收藏数量
	IsDel       int32     `gorm:"column:is_del;not null" json:"-"`                                   // 是否删除,0未删除，1已删除
	Status       int     `gorm:"column:status;not null" json:"status"`                                   // 状态,0未上架，1已上架
	CollectFlag int       `gorm:"-" json:"collect_flag"`   //用户收藏标识
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"-"` // 更新时间
}

// TableName Guide's table name
func (*Guide) TableName() string {
	return TableNameGuide
}

func NewGuide() *Guide {
	return new(Guide)
}

//新增功能略
func (m *Guide) Create() error {
	return db.Save(m).Error
}

//FindGuideByGameId 根据游戏id查找评论数据信息
func (m *Guide) FindGuideByGameId(gameId int32, page, size, tab int) (guides []*Guide, count int64, err error) {
	offset := (page - 1) * size
	tx := db.Model(m).Select("id,title,`describe`,rate,url,author,visit,collect,created_at").Where("is_del = 0")
	if gameId != 0 {
		tx = tx.Where("game_id = ?", gameId)
	}
	if tab == 1 {
		tx = tx.Order("collect desc")
	} else {
		tx = tx.Order("created_at desc")
	}
	err = tx.Where("status = 1").Count(&count).Offset(offset).Limit(size).Find(&guides).Error
	return
}

//FindInfoById 根据id查找详情
func (m *Guide) FindInfoById(id int, isAdmin bool) (guide *Guide, err error) {
	tx := db.Model(m).Where("id = ?", id)
	if !isAdmin {
		tx = tx.Where("status = 1")
	}
	err = tx.Find(&guide).Error
	return
}

func (m *Guide) Update() error {
	return db.Save(m).Error
}

func (m *Guide) FindGuideAdminList(page, size, status int, keyword string) (guide []*Guide, count int64, err error) {
	offset := (page - 1) * size
	tx := db.Model(m)
	if keyword != "" {
		s := fmt.Sprintf("%s", "%")
		s = fmt.Sprintf("%s%s", s, keyword)
		s = fmt.Sprintf("%s%s", s, "%")
		tx = tx.Where("game_name like ?", s)
	}
	if status != 2 {
		tx = tx.Where("status = ?", status)
	}
	err = tx.Where("is_del = 0").Count(&count).Offset(offset).Limit(size).Order("created_at desc").Find(&guide).Error
	return
}

//UpdateGuideInfoByGuideId 根据攻略id更新数据信息
func (m *Guide) UpdateGuideInfoByGuideId(guideId int32, datas map[string]interface{}) (err error) {
	err = db.Model(m).Where(guideId).Updates(datas).Error
	return
}
