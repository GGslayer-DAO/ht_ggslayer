package service

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
	"strings"
)

const ChainImageUrl = "https://static.tokenview.io/icon/"

type GameService struct {
	Ctx        *gin.Context
	Tag        string
	Chain      string
	GameName   string
	StatusTag  string
	TimeTag    int
	SortTag    string
	SortColumn string
	IsGameTab  int
	UserId     int
}

type GameInfo struct {
	Id          int32               `json:"id"`
	CollectTag  int                 `json:"collect_tag"`
	Rank        int                 `json:"rank"`
	Rise        int                 `json:"rise"` //-1降，0持平，1上升
	Name        string              `json:"name"`
	GameName    string              `json:"game_name"`
	Logo        string              `json:"logo"`
	Describe    string              `json:"describe"`
	Motivate    int                 `json:"motivate"`
	StatusTag   string              `json:"status_tag"`
	Tags        []string            `json:"tags"`
	Chains      []string            `json:"chains"`
	ChainImages []map[string]string `json:"chain_images"`
	GameToken   []*model.GameToken  `json:"token"`
	GameOther   *GameOtherInfo      `json:"other"`
}

type GameOtherInfo struct {
	BalanceRate     string `json:"balance_rate"`      // 当前金额比率
	BalanceValue    string `json:"balance_value"`     // 当前价格
	ActiveUserRate  string `json:"active_user_rate"`  // 活跃用户比率
	ActiveUserValue int    `json:"active_user_value"` // 活跃用户数量
	SocialRate      string `json:"social_rate"`       // 社交分比率
	SocialValue     int    `json:"social_value"`      // 社交分数量
	HoldersRate     string `json:"holders_rate"`      // 持有用户比例
	HoldersValue    int    `json:"holders_value"`     // 持有用户数量
}

//GameRanking 游戏列表排名
func (s *GameService) GameRanking(page, size int) (gameInfos []*GameInfo, count int64, err error) {
	var allGameIds, tagGameIds, chainGameIds []int32
	eg := errgroup.Group{}
	if s.Tag != "" {
		//根据标签查找
		eg.Go(func() error {
			tagGameIds, err = model.NewGameTag().FindGameIdByTag(s.Tag)
			return err
		})
	}
	if s.Chain != "" {
		//根据链查找
		eg.Go(func() error {
			chainGameIds, err = model.NewGameBlockChain().FindGameIdByChain(s.Chain)
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("GameRanking error(%s)", err.Error())
		return
	}
	//取gameIds交集
	if len(tagGameIds) > 0 && len(chainGameIds) == 0 {
		allGameIds = tagGameIds
	}
	if len(chainGameIds) > 0 && len(tagGameIds) == 0 {
		allGameIds = chainGameIds
	}
	//去除重复的game_id
	if len(tagGameIds) > 0 && len(chainGameIds) > 0 {
		allGameIds = utils.Intersect(tagGameIds, chainGameIds)
	}
	games, count, err := model.NewGames().FindIndex(page, size, allGameIds, s.GameName, s.StatusTag,
		s.TimeTag, s.SortColumn, s.SortTag, s.IsGameTab)
	if err != nil {
		log.Get(s.Ctx).Errorf("GameRanking FindIndex error(%s)", err.Error())
		return
	}
	gameIds := make([]int32, 0, len(games))
	tagMaps := make(map[int32][]string)
	chainsMaps := make(map[int32][]string)
	tokenMaps := make(map[int32][]*model.GameToken)
	otherMaps := make(map[int32]*model.GameOther)
	collectMaps := make(map[int32]int)
	rankMaps := make(map[int32]*model.GameRanking)
	for _, game := range games {
		gameIds = append(gameIds, game.ID)
	}
	//获取标签
	eg.Go(func() error {
		tagMaps, err = model.NewGameTag().FindTagsByGameIds(gameIds)
		return err
	})
	//获取游戏token数据
	eg.Go(func() error {
		tokenMaps, err = model.NewGameToken().FindGameTokenByGameId(gameIds)
		return err
	})
	//获取游戏其他数据
	eg.Go(func() error {
		otherMaps, err = model.NewGameOther().FindGameOtherByGameId(gameIds)
		return err
	})
	//获取所有游戏链
	eg.Go(func() error {
		chainsMaps, err = model.NewGameBlockChain().FindChainByGameId(gameIds)
		return err
	})
	//获取用户收藏数据
	eg.Go(func() error {
		if s.UserId == 0 {
			return nil
		}
		collectMaps, err = model.NewUserGameCollect().FindGameCollectByUserId(s.UserId, gameIds)
		return err
	})
	//获取所有game_ranking
	eg.Go(func() error {
		rankMaps, err = model.NewGameRanking().FindAllRankingByGameIds(gameIds)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("GameRanking error(%s)", err.Error())
		return
	}
	gameInfos = make([]*GameInfo, 0, len(games))
	for _, game := range games {
		collectTag := 0
		if v, ok := collectMaps[game.ID]; ok {
			collectTag = v
		}
		gameOtherInfo := new(GameOtherInfo)
		if v, ok := otherMaps[game.ID]; ok {
			gameOtherInfo.HoldersRate = v.HoldersRate
			gameOtherInfo.HoldersValue = v.HoldersValue
			gameOtherInfo.BalanceRate = v.BalanceRate
			gameOtherInfo.BalanceValue = v.BalanceValue
			if s.TimeTag == 2 {
				gameOtherInfo.SocialRate = v.SocialRate7d
				gameOtherInfo.SocialValue = v.SocialValue7d
				gameOtherInfo.ActiveUserRate = v.ActiveUserRate7d
				gameOtherInfo.ActiveUserValue = v.ActiveUserValue7d
			} else if s.TimeTag == 3 {
				gameOtherInfo.SocialRate = v.SocialRate30d
				gameOtherInfo.SocialValue = v.SocialValue30d
				gameOtherInfo.ActiveUserRate = v.ActiveUserRate30d
				gameOtherInfo.ActiveUserValue = v.ActiveUserValue30d
			} else {
				gameOtherInfo.SocialRate = v.SocialRate
				gameOtherInfo.SocialValue = v.SocialValue
				gameOtherInfo.ActiveUserRate = v.ActiveUserRate
				gameOtherInfo.ActiveUserValue = v.ActiveUserValue
			}
		}
		rank := 0
		rise := 0
		rank, rise = s.GetRankRise(s.IsGameTab, rankMaps, game.ID)
		gameInfos = append(gameInfos, &GameInfo{
			Rank:        rank,
			Rise:        rise,
			Id:          game.ID,
			Name:        game.Name,
			GameName:    game.GameName,
			Logo:        game.Logo,
			Describe:    game.Describe,
			Motivate:    cast.ToInt(game.Motivate),
			StatusTag:   game.StatusTag,
			Tags:        tagMaps[game.ID],
			Chains:      chainsMaps[game.ID],
			GameToken:   tokenMaps[game.ID],
			GameOther:   gameOtherInfo,
			CollectTag:  collectTag,
			ChainImages: s.GetChainImage(chainsMaps[game.ID]),
		})
	}
	return
}

func (s *GameService) GetRankRise(isGameTab int, rankMaps map[int32]*model.GameRanking, gameId int32) (rank, rise int) {
	if v, ok := rankMaps[gameId]; ok {
		if isGameTab != 4 {
			rank = v.SocialRank
			if v.SocialRank > v.SocialRankOld {
				rise = 1
			} else if v.SocialRank < v.SocialRankOld {
				rise = -1
			}
		} else {
			rank = v.VoteRank
			if v.VoteRank > v.VoteRankOld {
				rise = 1
			} else if v.VoteRank < v.VoteRankOld {
				rise = -1
			}
		}
	}
	return
}

//GetAllGameTags 获取所有游戏标签
func (s *GameService) GetAllGameTags() (tags []string, err error) {
	tags, err = model.NewGameTag().FindAllTags()
	return
}

//GetAllGameChains 获取所有游戏链
func (s *GameService) GetAllGameChains() (chains []string, err error) {
	chains, err = model.NewGameBlockChain().FindAllChains()
	return
}

//GetChainImage
func (s *GameService) GetChainImage(chains []string) []map[string]string {
	chainMapImages := make([]map[string]string, 0, len(chains))
	for _, chain := range chains {
		ch := strings.ToLower(chain)
		imageUrl := fmt.Sprintf("%s%s.png", ChainImageUrl, ch)
		chainMapImages = append(chainMapImages, map[string]string{
			"name": chain,
			"url":  imageUrl,
		})
	}
	return chainMapImages
}
