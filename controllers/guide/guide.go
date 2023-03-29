package guide

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"math"
)

const GuideVisit = "guide_visit"
const GuideCollect = "guide_collect"

//GameGuideList 查找游戏攻略
func GameGuideList(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 6)
	tab := ginc.GetInt(c, "tab", 0)
	guides, count, err := model.NewGuide().FindGuideByGameId(gameId, page, size, tab)
	if err != nil {
		log.Get(c).Errorf("GameGuideList FindGuideByGameId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if count == 0 {
		res := utils.PageReturn(page, size, 0, 0, []interface{}{})
		ginc.Ok(c, res)
		return
	}
	userId := c.GetInt("user_id")
	guidesIds := make([]int32, 0, len(guides))
	//获取用户收藏标识
	for _, guide := range guides {
		guidesIds = append(guidesIds, guide.ID)
	}
	collectMap, _ := model.NewUserGuideCollect().FindGuideCollectByUserId(userId, guidesIds)
	for _, guide := range guides {
		collectFlag := 0
		if v, ok := collectMap[guide.ID]; ok {
			collectFlag = v
		}
		visit := getVisitAndCollect(guide.ID)
		guide.Visit = visit
		guide.CollectFlag = collectFlag
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), guides)
	ginc.Ok(c, res)
}

func getVisitAndCollect(guideId int32) (visit int) {
	che := cache.RedisClient
	vist, _ := che.HGet(GuideVisit, cast.ToString(guideId)).Result()
	visit = cast.ToInt(vist)
	return
}

//GuideDetail 攻略详情
func GuideDetail(c *gin.Context) {
	guideId := ginc.GetInt(c, "id", 0)
	if guideId <= 0 {
		ginc.Fail(c, "page not found", "400")
		return
	}
	userId := c.GetInt("user_id")
	collectMap := make(map[int32]int)
	guide, err := model.NewGuide().FindInfoById(guideId, false)
	if userId != 0 {
		collectMap, _ = model.NewUserGuideCollect().FindGuideCollectByUserId(userId, []int32{int32(guideId)})
	}
	if err != nil {
		log.Get(c).Errorf("GuideDetail FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	visit := getVisitAndCollect(guide.ID)
	guide.Visit = visit
	if v, ok := collectMap[guide.ID]; ok {
		guide.CollectFlag = v
	}
	//增加访问量
	go addGuideVisit(guideId)
	ginc.Ok(c, guide)
}

//CollectGuide 收藏攻略
func CollectGuide(c *gin.Context) {
	guideId := ginc.GetInt(c, "guide_id", 0)
	if guideId <= 0 {
		ginc.Fail(c, "please input guide_id", "400")
		return
	}
	userId := c.GetInt("user_id")
	//加分布式锁
	md5LockKey := utils.Md5V(fmt.Sprintf("collect_guide%d%d", userId, guideId))
	err := cache.Lock(md5LockKey, "1", 5)
	if err != nil {
		ginc.Fail(c, "Do not operate frequently", "400")
		return
	}
	defer cache.UnLock(md5LockKey)
	//校验攻略id是否存在
	guide, err := model.NewGuide().FindInfoById(guideId, false)
	if err != nil {
		log.Get(c).Errorf("CollectGame FindGameByCond error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if guide.ID == 0 {
		ginc.Fail(c, "the guides is not exists", "400")
		return
	}
	flag, isAdd, err := model.NewUserGuideCollect().CreateOrUpdate(guideId, userId)
	if err != nil {
		log.Get(c).Errorf("CollectGuide CreateOrUpdate error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"flag": flag,
	}
	//攻略收藏增加经验值
	service.GuideCollectAddExp(userId, 2, 10, isAdd)
	addGuideCollect(guide, flag)
	ginc.Ok(c, res)
}

//AddGuideVisit 增加攻略访问量,访问量保存在redis
func addGuideVisit(guideId int) {
	cache.RedisClient.HIncrBy(GuideVisit, cast.ToString(guideId), 1)
}

//addGuideCollect 增加收藏攻略数量,访问量保存在redis
func addGuideCollect(guide *model.Guide, flag int) {
	go func() {
		collect := guide.Collect
		if flag == 1 {
			collect = int(utils.Add(float64(collect), 1, 0))
		} else {
			collect = int(utils.Sub(float64(collect), 1, 0))
		}
		if collect < 0 {
			return
		}
		guide.Collect = collect
		guide.Update()
	}()
}
