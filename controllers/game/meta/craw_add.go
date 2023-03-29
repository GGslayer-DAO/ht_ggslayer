package meta

import (
	"encoding/json"
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"github.com/gin-gonic/gin"
	"time"
)

//CrawMetaGameAdd mymeta平台game抓取
func CrawMetaGameAdd(c *gin.Context) {
	//爬取游戏数据
	index := ginc.GetInt(c, "index")
	if index <= 0 {
		ginc.Fail(c, "please input index")
		return
	}
	output, outputToken, cmap, _ := BaseCraw(c, index)
	//新增游戏
	records := output.Result.Records
	gameTokens := make([]*model.GameToken, 0)
	gameTags := make([]*model.GameTag, 0)
	gameBlockChains := make([]*model.GameBlockChain, 0)
	gameOther := make([]*model.GameOther, 0)
	gameActiveUser := make([]*model.GameActiveUser, 0)
	gameSocials := make([]*model.GameSocial, 0)
	for k, record := range records {
		game := new(model.Game)
		game.GameName = record.GameName
		game.Name = record.Name
		game.Logo = record.Logo
		game.CreatedAt = time.Now()
		game.UpdatedAt = time.Now()
		gameId, _ := game.Create()
		//插入game_token
		tokens := outputToken.Result
		ts := tokens[k+1] //从第2个开始
		for _, t := range ts {
			gameTokens = append(gameTokens, &model.GameToken{
				GameID:        gameId,
				ChangeRate24h: t.ChangeRate24H,
				TokenHolder:   int64(t.Holders),
				TokenPrice:    t.Price,
				Logo:          t.Logo,
				Symbol:        t.Symbol,
			})
		}
		//批量插入tag
		for _, tag := range record.Tags {
			gameTags = append(gameTags, &model.GameTag{
				GameID: gameId,
				Tag:    tag,
			})
		}
		//批量插入blockchain
		for _, chain := range record.BlockChain {
			gameBlockChains = append(gameBlockChains, &model.GameBlockChain{
				GameID:     gameId,
				BlockChain: chain,
				ShortName:  cmap[chain],
			})
		}
		//插入游戏其他数据信息
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
		gameOther = append(gameOther, &model.GameOther{
			GameID:          gameId,
			BalanceRate:     record.Balance["rate"],
			BalanceValue:    record.Balance["value"],
			ActiveUserRate:  activeUserRate,
			ActiveUserValue: activeUserValue,
			HoldersRate:     holdersRate,
			HoldersValue:    holdersValue,
			SocialRate:      socialRate,
			SocialValue:     socialValue,
		})
		//插入game_active_user
		gameActiveUser = append(gameActiveUser, &model.GameActiveUser{
			GameID:          gameId,
			ActiveUserValue: activeUserValue,
		})
		//插入game_social
		gameSocials = append(gameSocials, &model.GameSocial{
			GameID:      gameId,
			SocialValue: socialValue,
		})
	}
	//批量插入gameToken
	model.NewGameToken().BatchInserch(gameTokens)
	//批量插入gameTag
	model.NewGameTag().BatchInserch(gameTags)
	//批量插入gameTag
	model.NewGameBlockChain().BatchInserch(gameBlockChains)
	//批量插入gameOther
	model.NewGameOther().BatchInserch(gameOther)
	//批量插入gameActiveUser
	model.NewGameActiveUser().BatchInserch(gameActiveUser)
	//批量插入gameSocial
	model.NewGameSocial().BatchInserch(gameSocials)
}

type ContractResult struct {
	Code int           `json:"code"`
	Data *ContractData `json:"data"`
}

type ContractData struct {
	Addresses []*ContractAddresses `json:"addresses"`
}

type ContractAddresses struct {
	Id      int    `json:"id"`
	Token   string `json:"token"`
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

//ContractAddressAdd 添加智能合约地址
func ContractAddressAdd(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	url := ""
	gameTokens, _ := model.NewGameToken().FindGameToken(page, size)
	for _, gameToken := range gameTokens {
		url = "https://degame.com/game-home/api/open/token/web/info/"
		url = fmt.Sprintf("%s%s", url, gameToken.Symbol)
		res, _ := utils.NetLibGetV2(url, nil)
		contract := &ContractResult{}
		json.Unmarshal(res, contract)
		if contract.Data == nil {
			continue
		}
		addresses := contract.Data.Addresses
		for _, v := range addresses {
			chain, _ := model.NewGameBlockChain().FindByCond(gameToken.GameID, v.Chain)
			if chain.ID != 0 {
				chain.ContactAddress = v.Address
				chain.Update()
			}
		}
	}
	ginc.Ok(c, "success")
}

//ChangeChainName 修改简称
func ChangeChainName(c *gin.Context) {
	key := utils.Md5V("chain_list")
	var storage []*ChainRecords
	cache.GetStruct(key, &storage)
	cmap := make(map[string]string)
	for _, sage := range storage {
		cmap[sage.Name] = sage.ShowName
	}
	chains, _ := model.NewGameBlockChain().GetAllChain()
	for _, chain := range chains {
		if v, ok := cmap[chain.BlockChain]; ok {
			chain.ShortName = v
		}
		chain.Update()
	}
}

type OutputChain struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Result  *OutputChainResult `json:"result"`
}

type OutputChainResult struct {
	Count        int             `json:"count"`
	ChainRecords []*ChainRecords `json:"records"`
}

//MetaGameMetaChain 爬取链上数据
func MetaGameMetaChain(c *gin.Context) {
	variables := "eWZpNm45ZXlKc2FXMXBkQ0k2TlRBc0ltOW1abk5sZENJNk1IMD1jMzlsenBr"
	url := "https://www.mymetadata.io/api/v1/meta/chain/list?variables=" + variables
	outputChain := &OutputChain{}
	res, _ := utils.NetLibGet(url)
	json.Unmarshal(res, outputChain)
	md5key := utils.Md5V("chain_list")
	cache.SaveStruct(md5key, outputChain.Result.ChainRecords, time.Second*72000)
}
