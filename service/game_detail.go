package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"strings"
	"time"
)

type GameDetailService struct {
	Ctx    *gin.Context
	GameId int32
	UserId int
}

func NewGameDetailService(c *gin.Context, gameId int32, userId int) *GameDetailService {
	return &GameDetailService{
		Ctx:    c,
		GameId: gameId,
		UserId: userId,
	}
}

//GameDetailInfo 游戏详情数据返回
type GameDetailInfo struct {
	Id             int32               `json:"id"`
	GameName       string              `json:"game_name"`
	Logo           string              `json:"logo"`
	Describe       string              `json:"describe"`
	StatusTag      string              `json:"status_tag"`
	Video          string              `json:"video"`
	Detail         string              `json:"detail"`
	DetailDesc     string              `json:"desc"`
	GameTerm       []map[string]string `json:"game_term"`
	GameInvestor   []map[string]string `json:"game_investor"`
	SocialLink     []map[string]string `json:"social_link"`
	Website        string              `json:"website"`
	Screenshot     []string            `json:"screenshot"`
	Platform       []string            `json:"platform"`
	BlockChain     []string            `json:"block_chain"`
	Tags           []string            `json:"tags"`
	GameRate       float64             `json:"game_rate"`
	CollectFlag    int                 `json:"collect_flag"`
	PickPersonPics []string            `json:"pick_person_pics"`
	Motivate       int                 `json:"motivate"`
	TokenMaps      []*model.GameToken  `json:"game_token"`
	OtherMaps      *GameOtherInfoV2    `json:"game_other"`
}

type GameOtherInfoV2 struct {
	BalanceRate     string  `json:"balance_rate"`      // 当前金额比率
	BalanceValue    string  `json:"balance_value"`     // 当前价格
	ActiveUserRate  string  `json:"active_user_rate"`  // 活跃用户比率
	ActiveUserValue int     `json:"active_user_value"` // 活跃用户数量
	HoldersRate     string  `json:"holders_rate"`      // 持有用户比例
	HoldersValue    int     `json:"holders_value"`     // 持有用户数量
	VolumeRate      string  `json:"volume_rate"`       //交易比例
	VolumeValue     float64 `json:"volume_value"`      //交易量
}

//FindDetailInfo 查找详情数据信息
func (s *GameDetailService) FindDetailInfo() (gameDetailInfo *GameDetailInfo, err error) {
	var gameInfo *model.Game
	var detailInfo *model.GameDetail
	var chains []string
	var pickPersonPics []string
	var tags []string
	collectFlag := 0
	tokenMaps := make(map[int32][]*model.GameToken)
	otherMaps := make(map[int32]*model.GameOther)
	eg := errgroup.Group{}
	eg.Go(func() error {
		gameInfo, err = model.NewGames().FindInfoByGameId(s.GameId, false)
		return err
	})
	eg.Go(func() error {
		detailInfo, err = model.NewGameDetail().FindDetailByGameId(s.GameId)
		return err
	})
	//获取链上数据
	eg.Go(func() error {
		chains, err = model.NewGameBlockChain().FindBlockChainByGameId(s.GameId)
		return err
	})
	//获取标签数据
	eg.Go(func() error {
		tags, err = model.NewGameTag().FindTagByGameId(s.GameId)
		return err
	})
	//获取游戏pick用户，前10个
	eg.Go(func() error {
		userIds, er := model.NewMyGameVote().FindVotePickForGame(s.GameId)
		if er != nil {
			return er
		}
		//获取用户头像
		pickPersonPics, err = model.NewUser().FindHeadPicByUserId(userIds)
		return err
	})
	//判断是否有无收藏
	eg.Go(func() error {
		if s.UserId == 0 {
			return nil
		}
		collectFlag, err = model.NewUserGameCollect().FindByUserIdAndGameId(s.UserId, s.GameId)
		return err
	})
	//获取游戏token数据
	eg.Go(func() error {
		tokenMaps, err = model.NewGameToken().FindGameTokenByGameId([]int32{s.GameId})
		return err
	})
	//获取game_other数据
	eg.Go(func() error {
		otherMaps, err = model.NewGameOther().FindGameOtherByGameId([]int32{s.GameId})
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("FindDetailInfo error(%s)", err.Error())
		return
	}
	gameTermMap := make([]map[string]string, 0)
	gameInvestor := make([]map[string]string, 0)
	socialLink := make([]map[string]string, 0)
	json.Unmarshal([]byte(detailInfo.GameTerm), &gameTermMap)
	json.Unmarshal([]byte(detailInfo.GameInvestor), &gameInvestor)
	json.Unmarshal([]byte(detailInfo.SocialLink), &socialLink)
	screenshot := strings.Split(detailInfo.Screenshot, ",")
	platform := strings.Split(detailInfo.Platform, ",")
	otherV2 := &GameOtherInfoV2{}
	if other, ok := otherMaps[gameInfo.ID]; ok {
		otherV2.VolumeValue = other.VolumeValue
		otherV2.VolumeRate = other.VolumeRate
		otherV2.ActiveUserValue = other.ActiveUserValue
		otherV2.ActiveUserRate = other.ActiveUserRate
		otherV2.HoldersValue = other.HoldersValue
		otherV2.HoldersRate = other.HoldersRate
		otherV2.BalanceRate = other.BalanceRate
		otherV2.BalanceRate = other.BalanceRate
	}
	gameDetailInfo = &GameDetailInfo{
		Id:             gameInfo.ID,
		GameName:       gameInfo.GameName,
		Logo:           gameInfo.Logo,
		Describe:       gameInfo.Describe,
		StatusTag:      gameInfo.StatusTag,
		Video:          detailInfo.Video,
		Detail:         detailInfo.Detail,
		DetailDesc:     detailInfo.Describe,
		GameTerm:       gameTermMap,
		GameInvestor:   gameInvestor,
		SocialLink:     socialLink,
		Website:        detailInfo.Website,
		Screenshot:     screenshot,
		Platform:       platform,
		BlockChain:     chains,
		Tags:           tags,
		CollectFlag:    collectFlag,
		Motivate:       int(gameInfo.Motivate),
		PickPersonPics: pickPersonPics,
		TokenMaps:      tokenMaps[gameInfo.ID],
		OtherMaps:      otherV2,
	}
	return
}

type KolResult struct {
	Count   int         `json:"count"`
	Records []*KolTrust `json:"records"`
}

type KolTrust struct {
	Nick           string   `json:"nick"`
	Tid            string   `json:"tid"`
	Avatar         string   `json:"avatar"`
	FollowersCount int      `json:"followers_count"`
	ContentType    []string `json:"content_type"`
}

func (s *GameDetailService) GetCacheGameKolInfo() (kolResult *KolResult, err error) {
	storage := CacheStorage{Data: &kolResult}
	key := utils.Md5V(fmt.Sprintf("meta_game+%d", s.GameId))
	err = cache.GetStruct(key, &storage)
	if err == nil {
		if time.Now().Unix()-storage.TimeStamp > 24*3600 {
			go s.GetGameKolInfo(key)
		}
		return
	}
	return s.GetGameKolInfo(key)
}

func (s *GameDetailService) GetGameKolInfo(key string) (kolResult *KolResult, err error) {
	//查询游戏数据信息
	game, err := model.NewGames().FindInfoByGameId(s.GameId, false)
	if err != nil {
		log.Get(s.Ctx).Errorf("GetGameKolInfo FindInfoByGameId error(%s)", err.Error())
		return
	}
	gameName := game.GameName
	params := map[string]interface{}{
		"query_name": gameName,
		"is_all":     1,
	}
	url := "https://www.mymetadata.io/api/v1/meta/game/kol/followers?variables="
	variables := utils.MetaEncode(params)
	url = fmt.Sprintf("%s%s", url, variables)
	getRes, err := utils.NetLibGetV3(url, nil)
	if err != nil {
		log.Get(s.Ctx).Errorf("GetGameKolInfo NetLibGetV3 error(%s)", err.Error())
		return
	}
	kolResult = &KolResult{}
	res := Output{Result: kolResult}
	err = json.Unmarshal(getRes, &res)
	//存储redis
	storage := &CacheStorage{}
	storage.Data = kolResult
	storage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, storage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}
