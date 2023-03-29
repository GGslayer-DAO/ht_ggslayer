package meta

import (
	"ggslayer/model"
	"ggslayer/utils/ginc"
	"github.com/gin-gonic/gin"
	"time"
)

//CrawMetaGameUpdate mymeta平台game抓取更新
func CrawMetaGameUpdate(c *gin.Context) {
	//爬取游戏数据
	index := ginc.GetInt(c, "index", 0)
	if index <= -1 {
		ginc.Fail(c, "please input index")
		return
	}
	output, outputToken, _, _ := BaseCraw(c, index)
	//更新游戏
	records := output.Result.Records
	gameActiveUser := make([]*model.GameActiveUser, 0)
	gameSocials := make([]*model.GameSocial, 0)
	for k, record := range records {
		game, _ := model.NewGames().FindInfoByName(record.GameName)
		if game.ID == 0 {
			continue
		}
		game.GameName = record.GameName
		game.Name = record.Name
		game.Logo = record.Logo
		game.UpdatedAt = time.Now()
		game.Update()
		//更新game_token
		tokens := outputToken.Result
		ts := tokens[k+1] //从第2个开始
		for _, t := range ts {
			gameToken := &model.GameToken{
				Logo:          t.Logo,
				TokenPrice:    t.Price,
				TokenHolder:   int64(t.Holders),
				ChangeRate24h: t.ChangeRate24H,
				UpdatedAt:     time.Now(),
			}
			gameToken.UpdateGameTokenByGameId(game.ID, t.Symbol)
		}
		//更新game_other
		activeUserRate := ""
		activeUserValue := 0
		if record.ActiveUser != nil {
			activeUserRate = record.ActiveUser.Rate
			activeUserValue = record.ActiveUser.Value
		}
		holdersRate := ""
		holdersValue := 0
		if record.Holders != nil {
			holdersRate = record.Holders.Rate
			holdersValue = record.Holders.Value
		}
		socialRate := ""
		socialValue := 0
		if record.Social != nil {
			socialRate = record.Social.Rate
			socialValue = record.Social.Value
		}
		gameOther := &model.GameOther{
			BalanceRate:     record.Balance["rate"],
			BalanceValue:    record.Balance["value"],
			ActiveUserRate:  activeUserRate,
			ActiveUserValue: activeUserValue,
			SocialRate:      socialRate,
			SocialValue:     socialValue,
			HoldersRate:     holdersRate,
			HoldersValue:    holdersValue,
		}
		gameOther.UpdateGameOtherByGameId(int(game.ID))
		//插入game_active_user
		gameActiveUser = append(gameActiveUser, &model.GameActiveUser{
			GameID:          game.ID,
			ActiveUserValue: activeUserValue,
		})
		//插入game_social
		gameSocials = append(gameSocials, &model.GameSocial{
			GameID:      game.ID,
			SocialValue: socialValue,
		})
	}
	//批量插入gameActiveUser
	model.NewGameActiveUser().BatchInserch(gameActiveUser)
	//批量插入gameSocial
	model.NewGameSocial().BatchInserch(gameSocials)
}
