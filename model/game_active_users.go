// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"bytes"
	"fmt"
	"ggslayer/utils"
	"time"
)

const TableNameGameActiveUser = "game_active_users"

// GameActiveUser mapped from table <game_active_users>
type GameActiveUser struct {
	ID              int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                      // 主键ID
	GameID          int32     `gorm:"column:game_id;not null" json:"game_id"`                                 // 游戏id
	ActiveUserValue int   `gorm:"column:active_user_value;not null" json:"active_user_value"`             // 活跃用户数量
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
}

func NewGameActiveUser() *GameActiveUser {
	return &GameActiveUser{}
}

// TableName GameActiveUser's table name
func (*GameActiveUser) TableName() string {
	return TableNameGameActiveUser
}

const (
	_insertGameActiveUserSql = "insert into game_active_users(game_id,active_user_value,created_at) values "
)

//BatchInserch 批量插入
func (m *GameActiveUser) BatchInserch(gameActiveUsers []*GameActiveUser) {
	db.Exec(m.GenInsertSQL(gameActiveUsers))
}

// GenInsertSQL 生成批量插入的SQL
func (m *GameActiveUser) GenInsertSQL(gameActiveUsers []*GameActiveUser) string {
	if len(gameActiveUsers) == 0 {
		return ""
	}
	var (
		buf bytes.Buffer
		sql string
		now = time.Now().Local().Format(utils.TimeLayoutStr)
	)
	buf.WriteString(_insertGameActiveUserSql)
	for _, v := range gameActiveUsers {
		s := fmt.Sprintf("(%d,%d,'%s'),", v.GameID, v.ActiveUserValue, now)
		buf.WriteString(s)
	}
	sql = buf.String()
	return sql[0 : len(sql)-1]
}

//FindByGameId 根据条件查询游戏活跃用户数量
func (m *GameActiveUser) FindByGameId(gameId int32, b, d, n int) (activeUsers []*GameActiveUser, err error) {
	startTime := ""
	endTime := ""
	if b == 1 {
		startTime = utils.DateFt(d, 2 * n)
		endTime = utils.DateFt(d, n)
	} else {
		startTime = utils.DateFt(d, n)
		endTime = time.Now().Format(utils.TimeLayoutStr)
	}
	err = db.Model(m).Where("game_id = ?", gameId).
		Where("created_at >= ? and created_at <= ?", startTime, endTime).
		Find(&activeUsers).Error
	return
}

//CalcRate计算比例
func (m *GameActiveUser) CalcRateByGameId(gameId int32) (bv7d, bv1m int, rate7d, rate1m string) {
	bv7d = 0
	av7d := 0
	rate7d = ""
	//查询7天的game_active_user
	b7d, _ := m.FindByGameId(gameId, 0, 1, 7)
	a7d, _ := m.FindByGameId(gameId, 1, 1, 7)
	for _, b7 := range b7d {
		bv7d = bv7d + b7.ActiveUserValue
	}
	for _, a7 := range a7d {
		av7d = av7d + a7.ActiveUserValue
	}
	if av7d == 0 {
		rate7d = "100"
	} else {
		rate := utils.Div((float64(bv7d) - float64(av7d)), float64(av7d), 2)
		rate7d = fmt.Sprintf("%f", rate)
	}
	bv1m = 0
	av1m := 0
	rate1m = ""
	b1ms, _ := m.FindByGameId(gameId, 0, 2, 1)
	a1ms, _ := m.FindByGameId(gameId, 1, 2, 1)
	for _, b1m := range b1ms {
		bv1m = bv1m + b1m.ActiveUserValue
	}
	for _, a1m := range a1ms {
		av1m = av1m + a1m.ActiveUserValue
	}
	if av1m == 0 {
		rate1m = "100"
	} else {
		rate := utils.Div((float64(bv1m) - float64(av1m)), float64(av1m), 2)
		rate1m = fmt.Sprintf("%f", rate)
	}
	return
}
