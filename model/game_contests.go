// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameGameContest = "game_contests"
const GameContestVoteCache = "game_contest:"

// GameContest mapped from table <game_contests>
type GameContest struct {
	ID        int32 `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"` // 主键ID
	ContestID int32 `gorm:"column:contest_id;not null" json:"contest_id"`      // 比赛id
	GameID    int32 `gorm:"column:game_id;not null" json:"game_id"`            // 游戏id
	Vote      int `gorm:"column:vote;not null" json:"vote"`                  // 投票数量
	Rank      int `gorm:"column:rank;not null" json:"rank"`                  // 往期排名
}

// TableName GameContest's table name
func (*GameContest) TableName() string {
	return TableNameGameContest
}

func NewGameContest() *GameContest {
	return new(GameContest)
}

//FindGameContestInfoByContestId 根据contestId查找相关数据信息
func (m *GameContest) FindGameIdByContestId(contestId int32) (gameIds []int32, err error) {
	err = db.Model(m).Where("contest_id = ?", contestId).Pluck("game_id", &gameIds).Error
	return
}

//FindGameContestMapByContestId
func (m *GameContest) FindGameContestMapByContestId(contestId int32) (res map[int32]*GameContest, gameIds []int32, err error) {
	var gameContests []*GameContest
	err = db.Model(m).Where("contest_id = ?", contestId).Find(&gameContests).Error
	if err != nil {
		return
	}
	gameIds = make([]int32, 0, len(gameContests))
	res = make(map[int32]*GameContest)
	for _, gameContest := range gameContests {
		res[gameContest.GameID] = gameContest
		gameIds = append(gameIds, gameContest.GameID)
	}
	return
}

//FindByContestAndGameId 根据比赛和游戏id查找
func (m *GameContest) FindByContestAndGameId(contestId, gameId int32) (res *GameContest, err error) {
	err = db.Model(m).Where("contest_id = ? and game_id = ?", contestId, gameId).Find(&res).Error
	return
}

func (m *GameContest) UpdateByContestAndGameId(contestId, gameId int32, datas map[string]interface{}) (err error) {
	err = db.Model(m).Where("contest_id = ? and game_id = ?", contestId, gameId).Updates(datas).Error
	return
}

