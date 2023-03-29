// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"ggslayer/utils"
	"time"
)

const TableNameGameVote = "game_votes"

// GameVote mapped from table <game_votes>
type GameVote struct {
	ID       int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"` // 主键ID
	GameID   int32     `gorm:"column:game_id;not null" json:"game_id"`            // 游戏id
	Motivate int64     `gorm:"column:motivate" json:"motivate"`                   // 游戏当天票数
	Dates    time.Time `gorm:"column:dates;not null" json:"dates"`                // 时间
}

// TableName GameVote's table name
func (*GameVote) TableName() string {
	return TableNameGameVote
}

func NewGameVote() *GameVote {
	return &GameVote{}
}

//FirstOrCreate 游戏投票数据
func (m *GameVote) FirstOrCreate(gameId int32, motivate int64) (err error) {
	timeFormat := time.Now().Format("2006-01-02")
	vote := &GameVote{}
	db.Model(m).Where("game_id = ? and dates = ?", gameId, timeFormat).Find(vote)
	if vote.ID == 0 {
		vote.Dates = time.Now()
		vote.GameID = gameId
		vote.Motivate = motivate
	} else {
		voteNumber := int(utils.Add(float64(vote.Motivate), float64(motivate), 0))
		vote.Motivate = int64(voteNumber)
	}
	err = vote.Save()
	return
}

func (m *GameVote) Save() error {
	return db.Save(m).Error
}

type GameMotivate struct {
	GameId   int32
	Cvote    int
}

func (m *GameVote) FindGameVoteOnTop(t int) (gameMotivates []*GameMotivate, err error) {
	startTime := ""
	if t == 1 {
		startTime = utils.DateFt(1, 1)
	} else if t == 2 {
		startTime = utils.DateFt(1, 7)
	} else if t == 3 {
		startTime = utils.DateFt(2, 1)
	} else {
		startTime = "2021-12-32 00:00:00"
	}
	err = db.Model(m).Select("game_id, sum(motivate) as cvote").Where("dates >= ?", startTime).Group("game_id").
		Order("cvote desc").Limit(5).Find(&gameMotivates).Error
	return
}