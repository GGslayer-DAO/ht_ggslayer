package admin

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"math"
	"time"
)

//ContestList 比赛列表
func ContestList(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	status := ginc.GetInt(c, "status", 2)
	keyword := ginc.GetString(c, "keyword")
	contests, count, err := model.NewContest().FindContestAdminList(page, size, status, keyword)
	if err != nil {
		log.Get(c).Errorf("ContestList FindContestAdminList error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), contests)
	ginc.Ok(c, res)
}

//ContestCreate 比赛新增
func ContestCreate(c *gin.Context) {
	var v validate.ContestSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	if len(v.GameArr) != 10 {
		ginc.Fail(c, "game_arr is 10", "400")
		return
	}
	tx := model.GetDB().Begin()
	//新增比赛
	contest := &model.Contest{}
	contest.Title = v.Title
	contest.StartTime, _ = time.ParseInLocation(utils.TimeLayoutDate, v.StartTime, time.Local)
	contest.EndTime, _ = time.ParseInLocation(utils.TimeLayoutDate, v.EndTime, time.Local)
	err := tx.Save(contest).Error
	if err != nil {
		log.Get(c).Errorf("ContestCreate Save error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	contestId := contest.ID
	datas := make([]*model.GameContest, 0, len(v.GameArr))
	for _, gameId := range v.GameArr {
		datas = append(datas, &model.GameContest{
			ContestID: contestId,
			GameID:    gameId,
		})
	}
	err = tx.Create(&datas).Error
	if err != nil {
		log.Get(c).Errorf("ContestCreate Create error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx.Commit()
	AddRank(contestId, contest.Title, v.GameArr)
	ginc.Ok(c, "success")
}

//ContestDelete 比赛删除
func ContestDelete(c *gin.Context) {
	contestId := ginc.GetInt32(c, "contest_id", 0)
	if contestId <= 0 {
		ginc.Fail(c, "please input contest_id", "400")
		return
	}
	datas := map[string]interface{}{
		"is_del": 1,
	}
	err := model.NewContest().UpdateContestInfoByContestId(contestId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//ContestStatus 比赛状态变更
func ContestStatus(c *gin.Context) {
	contestId := ginc.GetInt32(c, "contest_id", 0)
	if contestId <= 0 {
		ginc.Fail(c, "please input contest_id", "400")
		return
	}
	status := ginc.GetInt(c, "status", 0)
	datas := map[string]interface{}{
		"status": status,
	}
	err := model.NewContest().UpdateContestInfoByContestId(contestId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//ContestEdit 比赛编辑页
func ContestEdit(c *gin.Context) {
	contestId := ginc.GetInt32(c, "id", 0)
	if contestId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	contest, err := model.NewContest().FindInfoById(contestId)
	if err != nil {
		log.Get(c).Errorf("ContestEdit FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	gameIds, err := model.NewGameContest().FindGameIdByContestId(contestId)
	if err != nil {
		log.Get(c).Errorf("ContestEdit FindGameIdByContestId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	contest.GameArr = gameIds
	ginc.Ok(c, contest)
}

//ContestUpdate 比赛更新
func ContestUpdate(c *gin.Context) {
	contestId := ginc.GetInt32(c, "id", 0)
	if contestId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	var v validate.ContestSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	if len(v.GameArr) != 10 {
		ginc.Fail(c, "game_arr is 10", "400")
		return
	}
	//比赛更新
	contest, err := model.NewContest().FindInfoById(contestId)
	if err != nil {
		log.Get(c).Errorf("EventUpdate FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx := model.GetDB().Begin()
	contest.Title = v.Title
	contest.StartTime, _ = time.ParseInLocation(utils.TimeLayoutDate, v.StartTime, time.Local)
	contest.EndTime, _ = time.ParseInLocation(utils.TimeLayoutDate, v.EndTime, time.Local)
	err = tx.Save(contest).Error
	if err != nil {
		log.Get(c).Errorf("ContestUpdate Save error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//先删除游戏比赛，再增加
	err = tx.Where("contest_id = ?", contestId).Delete(&model.GameContest{}).Error
	if err != nil {
		log.Get(c).Errorf("ContestUpdate gameContests DeleteByContestId error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	datas := make([]*model.GameContest, 0, len(v.GameArr))
	for _, gameId := range v.GameArr {
		datas = append(datas, &model.GameContest{
			ContestID: contestId,
			GameID:    gameId,
		})
	}
	err = tx.Create(&datas).Error
	tx.Commit()
	AddRank(contestId, contest.Title, v.GameArr)
	ginc.Ok(c, "success")
}

//AddRank 新增比赛排名
func AddRank(contestId int32, title string, gameIds []int32) {
	key := fmt.Sprintf("%s%d-%s", model.GameContestVoteCache, contestId, title)
	zarr := make([]redis.Z, 0, len(gameIds))
	for _, gameId := range gameIds {
		z := redis.Z{
			Score:  0,
			Member: gameId,
		}
		zarr = append(zarr, z)
	}
	cache.RedisClient.ZAdd(key, zarr...)
}
