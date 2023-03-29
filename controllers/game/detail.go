package game

import (
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"strings"
)

//GameDetail 获取游戏详情数据
func GameDetail(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	userId := c.GetInt("user_id")
	s := service.NewGameDetailService(c, gameId, userId)
	res, err := s.FindDetailInfo()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, res)
}

//KolGameInfo 获取kol数据(key：all, whale, platform, exchange, builder, vc, project, nft)
func KolGameInfo(c *gin.Context) {
	kol := ginc.GetString(c, "kol", "all")
	if !utils.InArray(kol, []string{"all", "whale", "platform", "exchange", "builder", "vc", "project", "nft"}) {
		ginc.Fail(c, "please input right kol key", "400")
		return
	}
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	s := service.NewGameDetailService(c, gameId, 0)
	kolInfo, err := s.GetCacheGameKolInfo()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	all, whale, platform, exchange, builder, vc, project, nft, kolTrusts := calKolInfo(kolInfo)
	res := map[string]interface{}{
		"all":      all,
		"whale":    whale,
		"platform": platform,
		"exchange": exchange,
		"builder":  builder,
		"vc":       vc,
		"project":  project,
		"nft":      nft,
	}
	if kol == "all" {
		res["list"] = kolInfo.Records
	} else {
		res["list"] = kolTrusts[kol]
	}
	ginc.Ok(c, res)
}

func calKolInfo(kolInfo *service.KolResult) (all, whale, platform, exchange, builder, vc, project, nft int, kolTrusts map[string][]*service.KolTrust) {
	all = kolInfo.Count
	kolTrusts = make(map[string][]*service.KolTrust)
	for _, record := range kolInfo.Records {
		for _, contentType := range record.ContentType {
			ct := strings.ToLower(contentType)
			switch ct {
			case "whale":
				whale++
			case "platform":
				platform++
			case "exchange":
				exchange++
			case "builder":
				builder++
			case "vc":
				vc++
			case "project":
				project++
			case "nft":
				nft++
			}
			//将key存储进去
			if v, ok := kolTrusts[ct]; ok {
				v = append(v, record)
				kolTrusts[ct] = v
			} else {
				kolTrusts[ct] = []*service.KolTrust{record}
			}
		}
	}
	return
}

func TokenInfo(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	s := service.NewGameDetailService(c, gameId, 0)
	tokenPrice, tokenHolders, activeUser, volume, err := s.GameTokenInfo()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	res := map[string]interface{}{
		"token_price":   tokenPrice,
		"token_holders": tokenHolders,
		"active_user":   activeUser,
		"volume":        volume,
	}
	ginc.Ok(c, res)
}

func NftInfo(c *gin.Context) {
	gameId := ginc.GetInt32(c, "game_id", 0)
	if gameId <= 0 {
		ginc.Fail(c, "please input game_id", "400")
		return
	}
	s := service.NewGameDetailService(c, gameId, 0)
	gameNfts, err := s.GetCacheGameNft()
	if err != nil {
		log.Get(c).Errorf("NftInfo FindGameNft error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, gameNfts)
}
