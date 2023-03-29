package admin

import (
	"encoding/json"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
	"math"
	"strings"
	"time"
)

//GameList 游戏列表
func GameList(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	status := ginc.GetInt(c, "status", 2)
	keyword := ginc.GetString(c, "keyword")
	games, count, err := model.NewGames().FindForAdminList(page, size, status, keyword)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), games)
	ginc.Ok(c, res)
}

//GameDelete 游戏删除
func GameDelete(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	datas := map[string]interface{}{
		"is_del": 1,
	}
	err := model.NewGames().UpdateGameInfoByGameId(gameId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GameStatus 游戏上下架
func GameStatus(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	status := ginc.GetInt(c, "status", 0)
	datas := map[string]interface{}{
		"status": status,
	}
	err := model.NewGames().UpdateGameInfoByGameId(gameId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GameCreate 新增游戏
func GameCreate(c *gin.Context) {
	var v validate.GameSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	tx := model.GetDB().Begin()
	//新增游戏数据
	game := model.NewGames()
	game.Name = v.Name
	game.GameName = v.GameName
	game.Logo = v.Logo
	game.StatusTag = v.StatusTag
	game.Describe = v.Describe
	game.UpdatedAt = time.Now()
	err := tx.Create(game).Error
	if err != nil {
		log.Get(c).Errorf("GameCreate game Create error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	gameId := game.ID
	//存储游戏详情数据
	detail := model.NewGameDetail()
	detail.Detail = v.Detail
	detail.GameID = gameId
	detail.Platform = strings.Join(v.Platform, ",")
	detail.Screenshot = strings.Join(v.Screenshot, ",")
	if len(v.SocialLink) > 0 {
		socialLink, _ := json.Marshal(v.SocialLink)
		detail.SocialLink = string(socialLink)
	}
	detail.Video = v.Video
	detail.Website = v.Website
	err = tx.Save(detail).Error
	if err != nil {
		log.Get(c).Errorf("GameCreate detail Save error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//存储标签数据
	if len(v.Tags) > 0 {
		var gameTags []*model.GameTag
		for _, tag := range v.Tags {
			gameTags = append(gameTags, &model.GameTag{
				GameID: gameId,
				Tag:    tag,
			})
		}
		err = tx.Create(gameTags).Error
		if err != nil {
			log.Get(c).Errorf("GameCreate tags Create error(%s)", err.Error())
			tx.Rollback()
			ginc.Fail(c, err.Error(), "400")
			return
		}
	}
	//存储链上数据
	if len(v.Chains) > 0 {
		var gameChains []*model.GameBlockChain
		for _, chain := range v.Chains {
			blockChain := new(model.GameBlockChain)
			blockChain.GameID = gameId
			if vc, ok := chain["chain"]; ok {
				blockChain.BlockChain = cast.ToString(vc)
				blockChain.ShortName = cast.ToString(vc)
			}
			if vc, ok := chain["symbol"]; ok {
				blockChain.Symbol = cast.ToString(vc)
			}
			if vc, ok := chain["contact_address"]; ok {
				blockChain.ContactAddress = cast.ToString(vc)
			}
			if vc, ok := chain["collections"]; ok {
				blockChain.Collection = cast.ToString(vc)
			}
			gameChains = append(gameChains, blockChain)
		}
		err = tx.Create(gameChains).Error
		if err != nil {
			log.Get(c).Errorf("GameCreate chains Create error(%s)", err.Error())
			tx.Rollback()
			ginc.Fail(c, err.Error(), "400")
			return
		}
	}
	tx.Commit()
	ginc.Ok(c, "success")
}

//GameTermAndInvestor 添加游戏管理员和投资人
func GameTermAndInvestor(c *gin.Context) {
	var v validate.GameInvestorValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	gameTerm, _ := json.Marshal(v.GameTerm)
	gameInvestor, _ := json.Marshal(v.GameInvestor)
	datas := map[string]interface{}{
		"game_term":     string(gameTerm),
		"game_investor": string(gameInvestor),
	}
	err := model.NewGameDetail().UpdateInfoByGameId(v.GameId, datas)
	if err != nil {
		log.Get(c).Errorf("GameTermAndInvestor UpdateInfoByGameId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

type GameEditInfo struct {
	ID         int32                   `json:"id"`
	Name       string                  `json:"name"`
	GameName   string                  `json:"game_name"`
	Logo       string                  `json:"logo"`
	Describe   string                  `json:"describe"`
	Motivate   int64                   `json:"motivate"`
	SocialRate int32                   `json:"social_rate"`
	StatusTag  string                  `json:"status_tag"`
	Detail     string                  `json:"detail"`
	Video      string                  `json:"video"`
	SocialLink []map[string]string     `json:"social_link"`
	Website    string                  `json:"website"`
	Screenshot []string                `json:"screenshot"`
	Platform   []string                `json:"platform"`
	Tags       []string                `json:"tags"`
	Chains     []*model.GameBlockChain `json:"chains"`
}

//GameEdit 游戏编辑
func GameEdit(c *gin.Context) {
	gameId := ginc.GetInt32(c, "id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	var (
		err        error
		game       *model.Game
		gameDetail *model.GameDetail
		gameTags   []string
		gameChains []*model.GameBlockChain
	)
	eg := errgroup.Group{}
	//获取游戏数据信息
	eg.Go(func() error {
		game, err = model.NewGames().FindInfoByGameId(gameId, true)
		return err
	})
	//获取游戏标签数据信息
	eg.Go(func() error {
		gameTags, err = model.NewGameTag().FindTagByGameId(gameId)
		return err
	})
	//获取游戏详情
	eg.Go(func() error {
		gameDetail, err = model.NewGameDetail().FindDetailByGameId(gameId)
		return err
	})
	//获取游戏链数据信息
	eg.Go(func() error {
		gameChains, err = model.NewGameBlockChain().FindChainInfoByGameId(gameId)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(c).Errorf("GameEdit error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	editInfo := &GameEditInfo{}
	editInfo.ID = game.ID
	editInfo.Name = game.Name
	editInfo.GameName = game.GameName
	editInfo.Logo = game.Logo
	editInfo.Describe = game.Describe
	editInfo.Motivate = game.Motivate
	editInfo.SocialRate = game.SocialRate
	editInfo.StatusTag = game.StatusTag
	editInfo.Video = gameDetail.Video
	editInfo.Detail = gameDetail.Detail
	editInfo.Platform = strings.Split(gameDetail.Platform, ",")
	editInfo.Website = gameDetail.Website
	socialLink := make([]map[string]string, 0)
	json.Unmarshal([]byte(gameDetail.SocialLink), &socialLink)
	screenshot := strings.Split(gameDetail.Screenshot, ",")
	editInfo.SocialLink = socialLink
	editInfo.Screenshot = screenshot
	editInfo.Tags = gameTags
	editInfo.Chains = gameChains
	ginc.Ok(c, editInfo)
}

//GameUpdate 游戏更新
func GameUpdate(c *gin.Context) {
	gameId := ginc.GetInt32(c, "id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	var v validate.GameSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	game, err := model.NewGames().FindInfoByGameId(gameId, true)
	if err != nil {
		log.Get(c).Errorf("GameUpdate FindInfoByGameId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx := model.GetDB().Begin()
	game.Name = v.Name
	game.GameName = v.GameName
	game.Logo = v.Logo
	game.StatusTag = v.StatusTag
	game.Describe = v.Describe
	err = tx.Updates(game).Error
	if err != nil {
		log.Get(c).Errorf("GameUpdate Updates error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//更新游戏详情数据
	detail := &model.GameDetail{}
	err = tx.Where("game_id = ?", gameId).Find(detail).Error
	if err != nil {
		log.Get(c).Errorf("GameUpdate FindDetailByGameId error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	detail.Detail = v.Detail
	detail.Platform = strings.Join(v.Platform, ",")
	detail.Screenshot = strings.Join(v.Screenshot, ",")
	if len(v.SocialLink) > 0 {
		socialLink, _ := json.Marshal(v.SocialLink)
		detail.SocialLink = string(socialLink)
	}
	detail.Video = v.Video
	detail.Website = v.Website
	err = tx.Updates(detail).Error
	if err != nil {
		log.Get(c).Errorf("GameUpdate detail Updates error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//更新游戏标签数据
	if len(v.Tags) > 0 {
		//先删除标签，再增加
		err = tx.Where("game_id = ?", gameId).Delete(&model.GameTag{}).Error
		if err != nil {
			log.Get(c).Errorf("GameUpdate gameTags DeleteByGameId error(%s)", err.Error())
			tx.Rollback()
			ginc.Fail(c, err.Error(), "400")
			return
		}
		var gameTags []*model.GameTag
		for _, tag := range v.Tags {
			gameTags = append(gameTags, &model.GameTag{
				GameID: gameId,
				Tag:    tag,
			})
		}
		err = tx.Create(gameTags).Error
		if err != nil {
			log.Get(c).Errorf("GameUpdate gameTags Create error(%s)", err.Error())
			tx.Rollback()
			ginc.Fail(c, err.Error(), "400")
			return
		}
	}
	//更新游戏链上数据
	if len(v.Chains) > 0 {
		//先删除再新增
		err = tx.Where("game_id = ?", gameId).Delete(&model.GameBlockChain{}).Error
		if err != nil {
			log.Get(c).Errorf("GameUpdate gameChains DeleteByGameId error(%s)", err.Error())
			tx.Rollback()
			ginc.Fail(c, err.Error(), "400")
			return
		}
		var gameChains []*model.GameBlockChain
		for _, chain := range v.Chains {
			blockChain := new(model.GameBlockChain)
			blockChain.GameID = gameId
			if vc, ok := chain["block_chain"]; ok {
				blockChain.BlockChain = cast.ToString(vc)
			}
			if vc, ok := chain["short_name"]; ok {
				blockChain.ShortName = strings.ToUpper(cast.ToString(vc))
			}
			if vc, ok := chain["symbol"]; ok {
				blockChain.Symbol = cast.ToString(vc)
			}
			if vc, ok := chain["contact_address"]; ok {
				blockChain.ContactAddress = cast.ToString(vc)
			}
			if vc, ok := chain["collections"]; ok {
				blockChain.Collection = cast.ToString(vc)
			}
			gameChains = append(gameChains, blockChain)
		}
		err = tx.Create(gameChains).Error
		if err != nil {
			log.Get(c).Errorf("GameUpdate chains Create error(%s)", err.Error())
			tx.Rollback()
			ginc.Fail(c, err.Error(), "400")
			return
		}
	}
	tx.Commit()
	ginc.Ok(c, "success")
}

//GameAddTrend 设置游戏最新趋势添加
func GameAddTrend(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	trend := ginc.GetInt(c, "trend", 1)
	num := ginc.GetInt(c, "num", 0)
	datas := make(map[string]interface{})
	if trend == 1 { //1是最新趋势
		datas["is_trending"] = num
	} else if trend == 2 { //2是最近的
		datas["is_recentry"] = num
	}
	err := model.NewGames().UpdateGameInfoByGameId(gameId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GameGetAll 查找所有游戏数据信息
func GameGetAll(c *gin.Context) {
	keyword := ginc.GetString(c, "keyword")
	res, err := model.NewGames().FindByNameByKeyword(keyword)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, res)
}
