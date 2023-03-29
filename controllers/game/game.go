package game

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"math"
)

//RankingGame 游戏排名
func RankingGame(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 30)
	tag := ginc.GetString(c, "tag", "")
	chain := ginc.GetString(c, "chain", "")
	gameName := ginc.GetString(c, "game_name", "")
	statusTag := ginc.GetString(c, "status_tag", "")
	timeTag := ginc.GetInt(c, "time_tag", 1)
	sortTag := ginc.GetString(c, "sort_tag", "")
	isGameTab := ginc.GetInt(c, "is_game_tab", 0)
	if !utils.InArray(sortTag, []string{"", "asc", "desc"}) {
		sortTag = "desc"
	}
	sortColumn := ginc.GetString(c, "sort_column", "")

	userId := c.GetInt("user_id")
	s := &service.GameService{
		Ctx:        c,
		Tag:        tag,
		Chain:      chain,
		GameName:   gameName,
		UserId:     userId,
		TimeTag:    timeTag,
		StatusTag:  statusTag,
		SortColumn: sortColumn,
		SortTag:    sortTag,
		IsGameTab:  isGameTab,
	}
	gameInfos, count, err := s.GameRanking(page, size)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := map[string]interface{}{
		"current_page": page,
		"next_page":    page + 1,
		"total":        count,
		"list":         gameInfos,
		"last_page":    lastPage,
		"size":         size,
	}
	ginc.Ok(c, res)
}

//AllTags 查找所有tag标签
func AllTags(c *gin.Context) {
	s := &service.GameService{
		Ctx: c,
	}
	tags, err := s.GetAllGameTags()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"tags": tags,
	}
	ginc.Ok(c, res)
}

//AllChains 查找所有游戏链
func AllChains(c *gin.Context) {
	s := &service.GameService{
		Ctx: c,
	}
	chains, err := s.GetAllGameChains()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"chains": chains,
	}
	ginc.Ok(c, res)
}

//CollectGame 游戏收藏
func CollectGame(c *gin.Context) {
	gameId := ginc.GetInt(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	userId := c.GetInt("user_id")
	//加分布式锁
	md5LockKey := utils.Md5V(fmt.Sprintf("collect_game%d%d", userId, gameId))
	err := cache.Lock(md5LockKey, "1", 5)
	if err != nil {
		ginc.Fail(c, "Do not operate frequently", "400")
		return
	}
	defer cache.UnLock(md5LockKey)
	//校验游戏id是否存在
	games, err := model.NewGames().FindGameByCond(map[string]interface{}{
		"game_id": gameId,
	})
	if err != nil {
		log.Get(c).Errorf("CollectGame FindGameByCond error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if len(games) == 0 {
		ginc.Fail(c, "the game_id is not exists", "400")
		return
	}
	flag, isAdd, err := model.NewUserGameCollect().CreateOrUpdate(gameId, userId)
	if err != nil {
		log.Get(c).Errorf("CollectGame CreateOrUpdate error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"flag": flag,
	}
	//游戏收藏增加经验值
	service.GameCollectAddExp(userId, 2, 10, isAdd)
	ginc.Ok(c, res)
}
