package service

import (
	"ggslayer/model"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

type HomeService struct {
	Ctx *gin.Context
}

func NewHomeService(c *gin.Context) *HomeService {
	return &HomeService{
		Ctx: c,
	}
}

type GameTopVote struct {
	Id        int32              `json:"id"`
	Rank      int                `json:"rank"`
	Rise      int                `json:"rise"` //-1降，0持平，1上升
	Name      string             `json:"name"`
	GameName  string             `json:"game_name"`
	Logo      string             `json:"logo"`
	Motivate  int                `json:"motivate"`
	StatusTag string             `json:"status_tag"`
	Tags      []string           `json:"tags"`
	Chains    []string           `json:"chains"`
	GameToken []*model.GameToken `json:"token"`
	GameOther *GameOtherInfo     `json:"other"`
}

//ShowHomeVoteGame 展示投票游戏
func (s *HomeService) ShowHomeVoteGame(timeTag int) (gameTopVoteInfos []*GameTopVote, err error) {
	cgameVotes, err := model.NewGameVote().FindGameVoteOnTop(timeTag)
	if err != nil {
		log.Get(s.Ctx).Errorf("ShowHomeGameVoteList FindGameVoteOnTop error(%s)", err.Error())
		return
	}
	gameVoteMap := make(map[int32]int, len(cgameVotes))
	gameIds := make([]int32, 0, len(cgameVotes))
	for _, cgameVote := range cgameVotes {
		gameIds = append(gameIds, cgameVote.GameId)
		gameVoteMap[cgameVote.GameId] = cgameVote.Cvote
	}
	var games []*model.Game
	tagMaps := make(map[int32][]string)
	chainsMaps := make(map[int32][]string)
	tokenMaps := make(map[int32][]*model.GameToken)
	otherMaps := make(map[int32]*model.GameOther)
	rankMaps := make(map[int32]*model.GameRanking)
	//根据game_id数组查询游戏信息
	eg := errgroup.Group{}
	eg.Go(func() error {
		games, err = model.NewGames().FindGameInfoByGameIds(gameIds)
		return err
	})
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
	//获取所有game_ranking
	eg.Go(func() error {
		rankMaps, err = model.NewGameRanking().FindAllRankingByGameIds(gameIds)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("ShowHomeVoteGame error(%s)", err.Error())
		return
	}
	gameTopVoteInfos = make([]*GameTopVote, 0, len(games))
	for _, game := range games {
		gameOtherInfo := new(GameOtherInfo)
		if v, ok := otherMaps[game.ID]; ok {
			gameOtherInfo.HoldersRate = v.HoldersRate
			gameOtherInfo.HoldersValue = v.HoldersValue
			gameOtherInfo.BalanceRate = v.BalanceRate
			gameOtherInfo.BalanceValue = v.BalanceValue
			if timeTag == 2 {
				gameOtherInfo.ActiveUserRate = v.ActiveUserRate7d
				gameOtherInfo.ActiveUserValue = v.ActiveUserValue7d
			} else if timeTag == 3 {
				gameOtherInfo.ActiveUserRate = v.ActiveUserRate30d
				gameOtherInfo.ActiveUserValue = v.ActiveUserValue30d
			} else {
				gameOtherInfo.ActiveUserRate = v.ActiveUserRate
				gameOtherInfo.ActiveUserValue = v.ActiveUserValue
			}
		}
		rank := 0
		rise := 0
		if v, ok := rankMaps[game.ID]; ok {
			rank = v.VoteRank
			if v.VoteRank > v.VoteRankOld {
				rise = 1
			} else if v.VoteRank < v.VoteRankOld {
				rise = -1
			}
		}
		gameTopVoteInfos = append(gameTopVoteInfos, &GameTopVote{
			Rank:      rank,
			Rise:      rise,
			Id:        game.ID,
			Name:      game.Name,
			GameName:  game.GameName,
			Logo:      game.Logo,
			Motivate:  cast.ToInt(gameVoteMap[game.ID]),
			StatusTag: game.StatusTag,
			Tags:      tagMaps[game.ID],
			Chains:    chainsMaps[game.ID],
			GameToken: tokenMaps[game.ID],
			GameOther: gameOtherInfo,
		})
	}
	return
}
