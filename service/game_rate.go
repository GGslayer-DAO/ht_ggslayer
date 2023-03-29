package service

import (
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"math"
	"time"
)

type GameRateService struct {
	UserId int
	GameId int32
	Ctx    *gin.Context
}

func NewGameRateService(c *gin.Context, userId int, gameId int32) *GameRateService {
	return &GameRateService{
		Ctx:    c,
		UserId: userId,
		GameId: gameId,
	}
}

//PostGameRate 发表评分评论，第一次有效
func (s *GameRateService) PostGameRate(rate int, experReason string) (err error) {
	//查询游戏数据信息
	_, err = model.NewGames().FindInfoByGameId(s.GameId, false)
	if err != nil {
		log.Get(s.Ctx).Errorf("PostGameRate FindInfoByGameId error(%s)", err.Error())
		return
	}
	//创建评论
	gameRate := &model.GameRate{}
	gameRate.GameID = s.GameId
	gameRate.UserID = int32(s.UserId)
	gameRate.Rate = rate
	gameRate.ExperReason = experReason
	gameRate.UpdatedAt = time.Now()
	err = gameRate.Create()
	if err != nil {
		log.Get(s.Ctx).Errorf("PostGameRate Create error(%s)", err.Error())
		return
	}
	//评论成功后续操作
	go s.RateToGame(rate)
	return
}

func (s *GameRateService) RateToGame(rate int) {
	//判断是否第一次评论
	rates, _ := model.NewGameRate().FindByUserIdAndGameId(s.GameId, int32(s.UserId))
	if len(rates) <= 1 {
		//增加评分有效数
		model.NewGameRateStatic().FirstOrCreate(s.GameId, rate)
	}
}

type UserRateInfo struct {
	ID          int32  `json:"id"`
	GameID      int32  `json:"game_id"`
	UserID      int32  `json:"user_id"`
	Rate        int    `json:"rate"`
	ExperReason string `json:"exper_reason"`
	UserName    string `json:"user_name"`
	UserPic     string `json:"user_pic"`
	CreatedAt   string `json:"created_at"`
	ThumbNumber int    `json:"thumb_number"`
	ThumbFlag   int    `json:"thumb_flag"`
}

//FindRateInfo 获取评论数据信息
func (s *GameRateService) FindRateInfo(page, size, tab int) (res map[string]interface{}, err error) {
	gameRates, count, err := model.NewGameRate().FindRateByGameId(s.GameId, page, size, tab)
	if err != nil {
		log.Get(s.Ctx).Errorf("GameRateList FindRateByGameId error(%s)", err.Error())
		return
	}
	if count == 0 {
		res = utils.PageReturn(page, size, 0, 0, []interface{}{})
		return
	}
	res, err = s.commonRateInfo(gameRates, count, page, size)
	return
}

func (s *GameRateService) commonRateInfo(gameRates []*model.GameRate, count int64, page, size int) (res map[string]interface{}, err error) {
	rateUserIds := make([]int32, 0, len(gameRates))
	rateIds := make([]int32, 0, len(gameRates))
	for _, gameRate := range gameRates {
		rateUserIds = append(rateUserIds, gameRate.UserID)
		rateIds = append(rateIds, gameRate.ID)
	}
	userMap := make(map[int32]*model.User)
	thumbMap := make(map[int32]int)
	eg := errgroup.Group{}
	eg.Go(func() error {
		userMap, err = model.NewUser().FindUserMapByUserIds(rateUserIds)
		return err
	})
	eg.Go(func() error {
		thumbMap, err = model.NewUserRateThumb().FindUserRateByUserId(s.UserId, rateIds)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("commonRateInfo error(%s)", err.Error())
		return
	}
	userRateInfos := make([]*UserRateInfo, 0, len(gameRates))
	for _, gameRate := range gameRates {
		thumbFlag := 0
		if v, ok := thumbMap[gameRate.ID]; ok {
			thumbFlag = v
		}
		userRateInfos = append(userRateInfos, &UserRateInfo{
			ID:          gameRate.ID,
			GameID:      gameRate.GameID,
			UserID:      gameRate.UserID,
			Rate:        gameRate.Rate,
			ExperReason: gameRate.ExperReason,
			UserName:    userMap[gameRate.UserID].Name,
			UserPic:     userMap[gameRate.UserID].HeadPic,
			ThumbNumber: gameRate.ThumbNumber,
			ThumbFlag:   thumbFlag,
			CreatedAt:   gameRate.CreatedAt.Format(utils.TimeLayoutStr),
		})
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res = utils.PageReturn(page, size, int(count), int(lastPage), userRateInfos)
	return
}

func (s *GameRateService) FindMyRateInfo(page, size int) (res map[string]interface{}, err error) {
	rates, count, err := model.NewGameRate().FindRateByUserId(s.UserId, page, size)
	if err != nil {
		log.Get(s.Ctx).Errorf("FindMyRateInfo FindRateByUserId error(%s)", err.Error())
		return
	}
	if count == 0 {
		res = utils.PageReturn(page, size, 0, 0, []interface{}{})
		return
	}
	res, err = s.commonRateInfo(rates, count, page, size)
	return
}
