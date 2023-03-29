package user

import (
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"strconv"
)

//BindAddressToken 绑定钱包地址
func BindAddressToken(c *gin.Context) {
	var v validate.BindTokenValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	userId := c.GetInt("user_id")
	tokenType := v.TokenType
	address := v.Address
	ut := model.NewUserToken().FindAddress(userId, tokenType, address)
	if ut.ID != 0 {
		ginc.Fail(c, "token address already exists", "400")
		return
	}
	ut.UserID = int32(userId)
	ut.Address = v.Address
	ut.TokenType = v.TokenType
	err := ut.Create()
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "bind user token address error", "400")
		return
	}
	//绑定钱包地址加30经验值
	service.AddExp(userId, 4, 30)
	ginc.Ok(c, "success")
}

//FindBalanceAssets 查询用户资产
func FindBalanceAssets(c *gin.Context) {
	address := ginc.GetString(c, "address")
	tokenType := ginc.GetString(c, "token_type", "token") //类型是token类型和nft类型
	if address == "" {
		ginc.Fail(c, "please input token address", "400")
		return
	}
	if !utils.InArray(tokenType, []string{"token", "nft"}) {
		ginc.Fail(c, "token_type must one of token or nft", "400")
		return
	}
	s := service.NewBalanceService(c, address)
	tob, nfts, err := s.GetCacheUserAddressBalance()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//代币游戏资产查询
	var tokenBill, nftBill, showBill []*service.AccessBill
	eg := errgroup.Group{}
	eg.Go(func() error {
		tokenBill, err = s.FindGameInfoByTob(tob)
		return err
	})
	eg.Go(func() error {
		nftBill, err = s.FindGameInfoByNft(nfts)
		return err
	})
	if err = eg.Wait(); err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	totalMoney := 0.0
	nftNumber := 0
	//计算token总金额
	for _, tb := range tokenBill {
		for _, t := range tb.Token {
			price, _ := strconv.ParseFloat(t.Price, 64)
			totalMoney = utils.Add(totalMoney, price, 2)
		}
	}
	if len(nftBill) == 0 {
		nftNumber = 0
	}
	if tokenType == "token" {
		showBill = tokenBill
	} else if tokenType == "nft" {
		showBill = nftBill
	}
	res := map[string]interface{}{
		"total_money": totalMoney,
		"nft_number":  nftNumber,
		"list":        showBill,
	}
	ginc.Ok(c, res)
}
