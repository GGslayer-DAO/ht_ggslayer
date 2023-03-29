package home

import (
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

//ShowHomeGameTab 首页展示游戏3个tag
func ShowHomeGameTab(c *gin.Context) {
	trendingGames := make([]*model.Game, 0, 3)
	recentryGames := make([]*model.Game, 0, 3)
	latestVcGames := make([]*model.Game, 0, 3)
	var err error
	eg := errgroup.Group{}
	//趋势
	condT := map[string]interface{}{
		"is_trending": 1,
		"limit":       3,
		"order":       "is_trending,desc",
	}
	eg.Go(func() error {
		trendingGames, err = model.NewGames().FindGameByCond(condT)
		return err
	})
	//最新的
	condR := map[string]interface{}{
		"is_recentry": 1,
		"limit":       3,
		"order":       "is_recentry,desc",
	}
	eg.Go(func() error {
		recentryGames, err = model.NewGames().FindGameByCond(condR)
		return err
	})
	//最新的空投投资
	condVc := map[string]interface{}{
		"is_latest_vc": 1,
		"limit":        3,
		"order":        "is_latest_vc,desc",
	}
	eg.Go(func() error {
		latestVcGames, err = model.NewGames().FindGameByCond(condVc)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(c).Errorf("ShowHomeGameTag error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"trending": trendingGames,
		"recentry": recentryGames,
		"latestVc": latestVcGames,
	}
	ginc.Ok(c, res)
}

//1是指24小时
//2是7天
//3是30天
//ShowHomeGameVoteList 展示首页游戏投票结果
func ShowHomeGameVoteList(c *gin.Context) {
	timeTag := ginc.GetInt(c, "time_tag", 1)
	s := service.NewHomeService(c)
	cgameVoteList, err := s.ShowHomeVoteGame(timeTag)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, cgameVoteList)
}
