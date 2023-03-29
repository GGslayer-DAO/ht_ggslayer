package contest

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
	"sort"
)

type GameContestInfo struct {
	GameId      int32    `json:"game_id"`
	GameName    string   `json:"game_name"`
	GameLogo    string   `json:"game_logo"`
	StatusTag   string   `json:"status_tag"`
	Tags        []string `json:"tags"`
	Chains      []string `json:"chains"`
	Vote        int      `json:"vote"`
	CollectFlag int      `json:"collect_flag"`
	Rank        int      `json:"rank"`
	Rise        int      `json:"rise"` //-1表示下降，0是平，1是升
}

//IndexInfo 展示比赛数据
func IndexInfo(c *gin.Context) {
	userId := c.GetInt("user_id")
	status := ginc.GetInt(c, "status", 1)
	title := ginc.GetString(c, "title")
	cond := make(map[string]interface{})
	cond["status"] = status
	if title != "" {
		cond["title"] = title
		cond["status"] = status
	}
	//查找比赛
	contest, err := model.NewContest().FindInfoByCond(cond)
	if err != nil {
		log.Get(c).Errorf("IndexInfo FindInfoByCond error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if contest.ID == 0 {
		ginc.Ok(c, []int{})
		return
	}
	//获取游戏id数组
	gameContestMap, gameIds, err := model.NewGameContest().FindGameContestMapByContestId(contest.ID)
	if err != nil {
		log.Get(c).Errorf("IndexInfo FindGameIdByContestId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	gameMap := make(map[int32]*model.Game)
	gameTagMap := make(map[int32][]string)
	gameChainMap := make(map[int32][]string)
	collectMaps := make(map[int32]int)
	var rankArr []redis.Z
	eg := errgroup.Group{}
	//获取游戏数据信息
	eg.Go(func() error {
		gameMap, err = model.NewGames().FindGameMapByIds(gameIds)
		return err
	})
	//获取标签数据
	eg.Go(func() error {
		gameTagMap, err = model.NewGameTag().FindTagsByGameIds(gameIds)
		return err
	})
	//获取游戏链数据信息
	eg.Go(func() error {
		gameChainMap, err = model.NewGameBlockChain().FindChainByGameId(gameIds)
		return err
	})
	//获取用户收藏标识
	eg.Go(func() error {
		if userId == 0 {
			return nil
		}
		collectMaps, err = model.NewUserGameCollect().FindGameCollectByUserId(userId, gameIds)
		return err
	})
	eg.Go(func() error {
		if status != 1 {
			return nil
		}
		key := fmt.Sprintf("%s%d-%s", model.GameContestVoteCache, contest.ID, contest.Title)
		rankArr, err = cache.RedisClient.ZRevRangeWithScores(key, 0, -1).Result()
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(c).Errorf("IndexInfo error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	voteMap := make(map[int32]int)
	rankMap := make(map[int32]int)
	for k, rz := range rankArr {
		voteMap[cast.ToInt32(rz.Member)] = cast.ToInt(rz.Score)
		rankMap[cast.ToInt32(rz.Member)] = k + 1
	}
	var gcs []*GameContestInfo
	for _, gameId := range gameIds {
		gc := &GameContestInfo{}
		gc.GameId = gameId
		if v, ok := gameMap[gameId]; ok {
			gc.GameName = v.GameName
			gc.GameLogo = v.Logo
			gc.StatusTag = v.StatusTag
		}
		if v, ok := gameTagMap[gameId]; ok {
			gc.Tags = v
		}
		if v, ok := gameChainMap[gameId]; ok {
			gc.Chains = v
		}
		if v, ok := collectMaps[gameId]; ok {
			gc.CollectFlag = v
		}
		rise := 0
		if v, ok := rankMap[gameId]; ok {
			gc.Rank = v
			if vc, vok := gameContestMap[gameId]; vok {
				if v > vc.Rank {
					rise = 1
				} else if v < vc.Rank {
					rise = -1
				}
			}
		}
		if v, ok := voteMap[gameId]; ok {
			gc.Vote = v
		}
		gc.Rise = rise
		gcs = append(gcs, gc)
	}
	sort.Slice(gcs, func(i, j int) bool {
		return gcs[i].Vote > gcs[j].Vote
	})
	ginc.Ok(c, gcs)
}

//IndexTitle 获取往期比赛标题
func IndexTitle(c *gin.Context) {
	res, err := model.NewContest().FindTitleByHistory()
	if err != nil {
		log.Get(c).Errorf("IndexTitle FindTitleByHistory error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, res)
}
