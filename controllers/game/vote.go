package game

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"sync"
	"time"
)

const GameVote = "game_vote:"
const GameVoteNumber = "game_vote_number:"
const GameForward = "game_forward:"
const GameForwardNumber = "game_forward_number:"

var Lock sync.Mutex

//VoteToGame 游戏投票
func VoteToGame(c *gin.Context) {
	vn, fn, err := VoteCommon(c, "vote", "vote_to_game", 1)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"vn": vn,
		"fn": fn,
	}
	ginc.Ok(c, res)
}

func ForwardToGame(c *gin.Context) {
	vn, fn, err := VoteCommon(c, "forward", "forward_to_game", 3)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"vn": vn,
		"fn": fn,
	}
	ginc.Ok(c, res)
}

func VoteCommon(c *gin.Context, method, lockName string, voteNumber int64) (vn, fn int, err error) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	userId := c.GetInt("user_id")
	if gameId <= 0 {
		err = fmt.Errorf("please input game_id")
		return
	}
	//加分布式锁
	md5LockKey := utils.Md5V(fmt.Sprintf("%s%d%d", lockName, userId, gameId))
	err = cache.Lock(md5LockKey, "1", 5)
	if err != nil {
		err = fmt.Errorf("do not operate frequently")
		return
	}
	defer cache.UnLock(md5LockKey)
	if method == "vote" {
		isCanVote := CheckVote(userId, int(gameId))
		if isCanVote == 0 {
			err = fmt.Errorf("can not to vote")
			return
		}
	} else {
		isCanForward := CheckForward(userId, int(gameId))
		if isCanForward == 0 {
			err = fmt.Errorf("can not to forward")
			return
		}
	}
	//校验游戏id是否存在
	game, err := model.NewGames().FindInfoByGameId(gameId, false)
	if game.ID == 0 {
		err = fmt.Errorf("the game_id is not exists")
		return
	}
	err = model.NewGameVote().FirstOrCreate(gameId, voteNumber)
	if err != nil {
		log.Get(c).Errorf("VoteToGame FirstOrCreate error(%s)", err.Error())
		return
	}
	vn, fn = ForVoteWardNumber(userId)
	if method == "vote" {
		vn = vn - 1
	} else {
		fn = fn - 1
	}
	VoteCache(userId, int(gameId), method)
	go VoteUpdate(game, int32(userId), voteNumber)
	return
}

//VoteCache 投票数据更新缓存
func VoteCache(userId, gameId int, method string) {
	key1 := ""
	key2 := ""
	ck1 := utils.Md5V(fmt.Sprintf("%d%d", userId, gameId))
	ck2 := utils.Md5V(fmt.Sprintf("%d", userId))
	if method == "vote" {
		key1 = fmt.Sprintf("%s%s", GameVote, ck1)
		key2 = fmt.Sprintf("%s%s", GameVoteNumber, ck2)
	} else {
		key1 = fmt.Sprintf("%s%s", GameForward, ck1)
		key2 = fmt.Sprintf("%s%s", GameForwardNumber, ck2)
	}
	client := cache.RedisClient
	tstr := "2006-01-02"
	client.Set(key1, "1", time.Hour*1)
	client.Incr(key2)
	t2, _ := time.ParseInLocation(tstr, time.Now().Format(tstr), time.Local)
	// 第二天零点过期
	tomollo := t2.AddDate(0, 0, 1)
	cache.RedisClient.ExpireAt(key1, tomollo)
	cache.RedisClient.ExpireAt(key2, tomollo)
}

//VoteUpdate 用户票数更新
func VoteUpdate(game *model.Game, userId int32, voteNumber int64) {
	Lock.Lock()
	defer Lock.Unlock()
	model.NewMyGameVote().FirstOrCreate(game.ID, userId, voteNumber) //用户票数更新
	motivate := game.Motivate
	game.Motivate = motivate + voteNumber
	game.Update() //游戏票数更新
	//校验是否在游戏比赛里
	cond := map[string]interface{}{
		"status": 1,
	}
	contest, _ := model.NewContest().FindInfoByCond(cond)
	if contest.ID == 0 {
		return
	}
	gameContest, _ := model.NewGameContest().FindByContestAndGameId(contest.ID, game.ID)
	if gameContest.ID != 0 {
		//新增排名
		key := fmt.Sprintf("%s%d-%s", model.GameContestVoteCache, contest.ID, contest.Title)
		cache.RedisClient.ZIncrBy(key, 1, cast.ToString(game.ID))
	}
}

//VoteIsCan 是否可以投票
func VoteIsCan(c *gin.Context) {
	gameId := ginc.GetInt(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	userId := c.GetInt("user_id")
	isCanVote := CheckVote(userId, gameId)
	isCanForward := CheckForward(userId, gameId)
	res := map[string]interface{}{
		"is_can_vote":    isCanVote,
		"is_can_forward": isCanForward,
	}
	ginc.Ok(c, res)
}

//CheckVote 校验投票
func CheckVote(userId, gameId int) (isCanVote int) {
	cacheClient := cache.RedisClient
	ck1 := utils.Md5V(fmt.Sprintf("%d%d", userId, gameId))
	ck2 := utils.Md5V(fmt.Sprintf("%d", userId))
	key1 := fmt.Sprintf("%s%s", GameVote, ck1)
	key2 := fmt.Sprintf("%s%s", GameVoteNumber, ck2)
	//判断是否存在redis，存在就不能投票
	isCanVote = 1
	isV, _ := cacheClient.Get(key1).Result()
	isVn, _ := cacheClient.Get(key2).Result()
	if isV != "" {
		isCanVote = 0
		return
	}
	if isVn == "" {
		return
	}
	isVnInt := cast.ToInt(isVn)
	if isVnInt >= 3 {
		isCanVote = 0
	}
	return
}

//CheckForward 校验转发
func CheckForward(userId, gameId int) (isCanForward int) {
	cacheClient := cache.RedisClient
	ck1 := utils.Md5V(fmt.Sprintf("%d%d", userId, gameId))
	ck2 := utils.Md5V(fmt.Sprintf("%d", userId))
	key1 := fmt.Sprintf("%s%s", GameForward, ck1)
	key2 := fmt.Sprintf("%s%s", GameForwardNumber, ck2)
	//判断是否存在redis，存在就不能投票
	isCanForward = 1
	isF, _ := cacheClient.Get(key1).Result()
	isFn, _ := cacheClient.Get(key2).Result()
	if isF != "" {
		isCanForward = 0
		return
	}
	if isFn == "" {
		return
	}
	isFnInt := cast.ToInt(isFn)
	if isFnInt >= 3 {
		isCanForward = 0
	}
	return
}

//ForVoteWardNumber 返回当天剩余投票数和转发数
func ForVoteWardNumber(userId int) (vn, fn int) {
	cacheClient := cache.RedisClient
	ck := utils.Md5V(fmt.Sprintf("%d", userId))
	key1 := fmt.Sprintf("%s%s", GameForwardNumber, ck)
	key2 := fmt.Sprintf("%s%s", GameVoteNumber, ck)
	isFn, _ := cacheClient.Get(key1).Result()
	isVn, _ := cacheClient.Get(key2).Result()
	if isFn == "" {
		fn = 3
	} else {
		fn = 3 - cast.ToInt(isFn)
	}
	if isVn == "" {
		vn = 3
	} else {
		vn = 3 - cast.ToInt(isVn)
	}
	return
}
