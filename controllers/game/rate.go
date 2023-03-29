package game

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

//GameRate 游戏评论
func GameRate(c *gin.Context) {
	var v validate.PostGameRateValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	userId := c.GetInt("user_id")
	s := service.NewGameRateService(c, userId, v.GameId)
	err := s.PostGameRate(v.Rate, v.ExperReason)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GameRateList 游戏评分展示
func GameRateList(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please must input game_id", "400")
		return
	}
	userId := c.GetInt("user_id")
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	tab := ginc.GetInt(c, "tab", 1) //1是热门，2是最新的
	s := service.NewGameRateService(c, userId, gameId)
	res, err := s.FindRateInfo(page, size, tab)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, res)
}

//RateThumb 评论点赞
func RateThumb(c *gin.Context) {
	rateId := ginc.GetInt(c, "rate_id", 0)
	if rateId <= 0 {
		ginc.Fail(c, "please must input rate_id", "400")
		return
	}
	userId := c.GetInt("user_id")
	//加分布式锁
	md5LockKey := utils.Md5V(fmt.Sprintf("rate_thumb%d%d", userId, rateId))
	err := cache.Lock(md5LockKey, "1", 5)
	if err != nil {
		ginc.Fail(c, "Do not operate frequently", "400")
		return
	}
	defer cache.UnLock(md5LockKey)
	//校验评论id是否存在
	rate, err := model.NewGameRate().FindRateById(rateId)
	if err != nil {
		log.Get(c).Errorf("RateThumb FindRateById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if rate.ID == 0 {
		ginc.Fail(c, "the rates is not exists", "400")
		return
	}
	flag, err := model.NewUserRateThumb().CreateOrUpdate(rateId, userId)
	if err != nil {
		log.Get(c).Errorf("RateThumb CreateOrUpdate error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"flag": flag,
	}
	//增加评论点赞数
	go addRateThumbNumber(rate, flag)
	ginc.Ok(c, res)
}

//增加评论点赞数
func addRateThumbNumber(rate *model.GameRate, flag int) {
	thumbNumber := 0.0
	if flag == 1 {
		thumbNumber = utils.Add(float64(rate.ThumbNumber), 1, 0)
	} else {
		thumbNumber = utils.Sub(float64(rate.ThumbNumber), 1, 0)
	}
	rate.ThumbNumber = cast.ToInt(thumbNumber)
	rate.Create()
}

//GameRateStatics 游戏评分数据统计
func GameRateStatics(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please must input game_id", "400")
		return
	}
	statics, err := model.NewGameRateStatic().FindByGameId(gameId)
	if err != nil {
		log.Get(c).Errorf("GameRateStatics FindByGameId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, statics)
}
