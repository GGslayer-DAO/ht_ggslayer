package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/model"
	"ggslayer/tasks"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/config"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"
)

const (
	Address_Account      = "address_account:"
	Address_Game_Account = "address_game_account"
	META_CACHE_TTL       = 259200 //元数据缓存时间(72小时)
)

var LockNft sync.Mutex

type BalanceService struct {
	Url            string
	Address        string
	ApiKey         string
	Tab            string
	Page           int
	CoinMarkUrl    string
	CoinMarkApiKey string
	Ctx            *gin.Context
}

func NewBalanceService(c *gin.Context, address string) *BalanceService {
	return &BalanceService{
		Address:        address,
		Url:            config.GetString("token.TokenUrl"),
		ApiKey:         config.GetString("token.TokenApiKey"),
		CoinMarkUrl:    config.GetString("token.CoinMarkUrl"),
		CoinMarkApiKey: config.GetString("token.CoinMarkApiKey"),
		Ctx:            c,
	}
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type CacheStorage struct {
	TimeStamp int64
	Data      interface{}
}

type AccessStorage struct {
	TimeStamp int64
	Token     interface{}
	Nft       interface{}
}

type AddressB struct {
	Type         string  `json:"type"`
	Network      string  `json:"network"`
	Hash         string  `json:"hash"`
	Balance      string  `json:"balance"`
	BalanceUnit  string  `json:"balance_unit"`
	Icon         string  `json:"icon"`
	UsdPrice     float64 `json:"usd_price"`
	UsdPriceUnit string  `json:"usd_price_unit"`
}

func (s *BalanceService) GetCacheAddressBalance() (tokenAllBt []*TokenBt, nftAllMeta []*NftMeta, err error) {
	key := utils.Md5V(fmt.Sprintf("%s%s%s", Address_Account, "address", s.Address))
	storage := &AccessStorage{
		Token: &tokenAllBt,
		Nft:   &nftAllMeta,
	}
	err = cache.GetStruct(key, &storage)
	if err == nil {
		if time.Now().Unix()-storage.TimeStamp > 600 {
			go s.FindAddressBalance(key)
		}
		return
	}
	return s.FindAddressBalance(key)
}

//FindAddressBalance 查询地址资产
func (s *BalanceService) FindAddressBalance(key string) (tokenAllBt []*TokenBt, nftAllMeta []*NftMeta, err error) {
	//获取地址所有链数据信息
	url := fmt.Sprintf("%ssearch/%s", s.Url, s.Address)
	getRes, err := utils.NetLibGetV3(url, nil)
	var addressB []*AddressB
	res := Result{Data: &addressB}
	err = json.Unmarshal(getRes, &res)
	if err != nil {
		log.Get(s.Ctx).Errorf("FindAddressBalance Unmarshal error(%s)", err.Error())
		return
	}
	if len(addressB) == 0 {
		return
	}
	lock := sync.Mutex{}
	eg := errgroup.Group{}
	for _, addrB := range addressB {
		newAddrB := addrB
		network := strings.ToLower(newAddrB.Network)
		//查询代币数据信息
		eg.Go(func() error {
			turl := fmt.Sprintf("https://%s.tokenview.io/api/%s/address/tokenbalance/%s", network, network, s.Address)
			tokenRes, er := utils.NetLibGetV3(turl, nil)
			if er != nil {
				return er
			}
			var tokenBt []*TokenBt
			rs := Result{Data: &tokenBt}
			err = json.Unmarshal(tokenRes, &rs)
			if err != nil {
				return err
			}
			lock.Lock()
			tokenAllBt = append(tokenAllBt, tokenBt...)
			lock.Unlock()
			return nil
		})
		//查询链上nft数据,默认查50条
		eg.Go(func() error {
			ns, er := s.FindBalanceNftByPage(network, 1)
			if er != nil {
				return er
			}
			nftMetas := ns.Data
			LockNft.Lock()
			nftAllMeta = append(nftAllMeta, nftMetas...)
			LockNft.Unlock()
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("FindAddressBalance error(%s)", err.Error())
	}
	accessStorage := &AccessStorage{}
	accessStorage.Token = tokenAllBt
	accessStorage.Nft = nftAllMeta
	accessStorage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, accessStorage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}

//GetAllNft 获取所有nft
func (s *BalanceService) GetAllNft(network string, num int) (NftAllMeta []*NftMeta, err error) {
	nurl := fmt.Sprintf("https://%s.tokenview.io/api/tokens/address/nft/all/%s/%s/%d/2", network, network, s.Address, num)
	nftRes, er := utils.NetLibGetV3(nurl, nil)
	if er != nil {
		return
	}
	//var nftMetas []*NftMeta
	nftData := &NftData{}
	rs := Result{Data: nftData}
	err = json.Unmarshal(nftRes, &rs)
	if err != nil {
		return
	}
	LockNft.Lock()
	NftAllMeta = append(NftAllMeta, nftData.Data...)
	LockNft.Unlock()
	return
}

type TokenBt struct {
	NetWork   string     `json:"network"`
	Hash      string     `json:"hash"`
	TokenInfo *TokenInfo `json:"tokenInfo"`
	Balance   string     `json:"balance"`
}

type TokenInfo struct {
	C string `json:"c"`
	D string `json:"d"`
	F string `json:"f"`
	H string `json:"h"`
	S string `json:"s"`
}

//GetCacheAddressToken 查找资产token
func (s *BalanceService) GetCacheAddressToken(network string) (tokenBt []*TokenBt, err error) {
	key := utils.Md5V(fmt.Sprintf("%s%s%s%s", Address_Account, "token", s.Address, network))
	storage := CacheStorage{Data: &tokenBt}
	err = cache.GetStruct(key, &storage)
	log.PP(storage)
	if err == nil {
		//todo
		return
	}
	return s.FindAddressToken(key, network)
}

//FindAddressToken 查找地址代币类型
func (s *BalanceService) FindAddressToken(key, network string) (newTokenBt []*TokenBt, err error) {
	network = strings.ToLower(network)
	url := fmt.Sprintf("%s%s/address/tokenbalance/%s?apikey=%s", s.Url, network, s.Address, s.ApiKey)
	getRes, err := utils.NetLibGetV1(url, nil)
	if err != nil {
		log.Get(s.Ctx).Error(err)
		return
	}
	var tokenBt []*TokenBt
	res := Result{Data: &tokenBt}
	err = json.Unmarshal(getRes, &res)
	if err != nil {
		log.Get(s.Ctx).Error(err)
		return
	}
	if len(tokenBt) == 0 {
		return
	}
	//把0余额要去掉
	newTokenBt = make([]*TokenBt, 0, len(tokenBt))
	for _, vt := range tokenBt {
		if vt.Balance == "0" {
			continue
		}
		if vt.TokenInfo.C != "" {
			continue
		}
		newTokenBt = append(newTokenBt, vt)
	}
	storage := &CacheStorage{}
	storage.Data = newTokenBt
	storage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, storage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}

//GetCacheAddressNft 查找资产nft
func (s *BalanceService) GetCacheAddressNft(network string, page int) (nft *NftData, err error) {
	key := utils.Md5V(fmt.Sprintf("%s%s%s%s%d", Address_Account, "nft", s.Address, network, page))
	storage := CacheStorage{Data: &nft}
	err = cache.GetStruct(key, &storage)
	if err == nil {
		//todo
		return
	}
	return s.FindAddressNft(key, network)
}

type NftData struct {
	Page  int        `json:"page"`
	Size  int        `json:"size"`
	Total int        `json:"total"`
	Data  []*NftMeta `json:"data"`
}

type NftMeta struct {
	Owner           string            `json:"owner"`
	Creator         string            `json:"creator"`
	TokenId         string            `json:"tokenId"`
	ImageUrl        string            `json:"imageUrl"`
	ContractAddress string            `json:"contractAddress"`
	OpenseaUrl      string            `json:"openseaUrl"`
	TokenInfo       map[string]string `json:"tokenInfo"`
}

//FindAddressNft 查找地址下面nft
func (s *BalanceService) FindAddressNft(key, network string) (nftData *NftData, err error) {
	network = strings.ToLower(network)
	url := fmt.Sprintf("https://%s.tokenview.io/api/tokens/address/nft/all/%s/%s/1/2", network, network, s.Address)
	getRes, err := utils.NetLibGetV3(url, nil)
	if err != nil {
		log.Get(s.Ctx).Error(err)
		return
	}
	nftData = &NftData{}
	res := Result{Data: nftData}
	err = json.Unmarshal(getRes, &res)
	if err != nil {
		log.Get(s.Ctx).Error(err)
		return
	}
	if len(nftData.Data) == 0 {
		return
	}
	storage := &CacheStorage{}
	storage.Data = nftData
	storage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, storage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}

//GetCacheRate 获取缓存币率
func (s *BalanceService) GetCacheRate() (rate map[string]float64, err error) {
	key := utils.Md5V(tasks.TOkEN_RATE)
	err = cache.GetStruct(key, &rate)
	return
}

type AccessBill struct {
	GameId      int32                  `json:"game_id"`
	GameName    string                 `json:"game_name"`
	Logo        string                 `json:"logo"`
	ChainAccess map[string]interface{} `json:"chain_access"`
	Token       []*TokenBill           `json:"token"`
	Nft         []*NftInfo             `json:"nft"`
	Chains      []string               `json:"chains"`
	ChainBill   []map[string]string    `json:"chain_bill"`
	GameRate    float64                `json:"game_rate"`
}

type TokenBill struct {
	Symbol         string `json:"symbol"`
	Logo           string `json:"logo"`
	Balance        string `json:"balance"`
	Price          string `json:"price"`
	Chain          string `json:"chain"`
	ContactAddress string `json:"contact_address"`
}

//FindGameInfoByAccess 代币游戏资产查询
func (s *BalanceService) FindGameInfoByAccess(tokenType string, tokenBt []*TokenBt, nftBt []*NftMeta) (accessBill []*AccessBill, err error) {
	//hashArr := make([]string, 0, len(tokenBt))
	//newTokenBt := make([]*TokenBt, 0, len(tokenBt))
	//symbolArr := make([]string, 0, len(tokenBt))
	//for _, token := range tokenBt {
	//	if tokenType == "token" {
	//		if token.TokenInfo.C != "" {
	//			continue
	//		}
	//	} else {
	//		if token.TokenInfo.C == "" {
	//			continue
	//		}
	//	}
	//	hashArr = append(hashArr, token.Hash)
	//	newTokenBt = append(newTokenBt, token)
	//	symbolArr = append(symbolArr, token.TokenInfo.S)
	//}
	//tokenMap := make(map[string]*model.GameToken)
	//var contactGameIds []int32
	//eg := errgroup.Group{}
	////eg.Go(func() error {
	////	tokenMap, err = model.NewGameToken().FindTokenBySymbol(symbolArr)
	////	return err
	////})
	//eg.Go(func() error {
	//	contactGameIds, _ = model.NewGameBlockChain().FindInfoByContactAddress(hashArr)
	//	return nil
	//})
	//if err = eg.Wait(); err != nil {
	//	log.Get(s.Ctx).Errorf("FindGameInfoByAccess error(%s)", err.Error())
	//	return
	//}
	////if len(tokenMap) == 0 {
	////	return
	////}
	//gameIds := make([]int32, 0, len(tokenMap))
	//newSymbolArr := make([]string, 0, len(tokenMap))
	//for _, token := range tokenMap {
	//	if utils.InArray(token.GameID, contactGameIds) {
	//		gameIds = append(gameIds, token.GameID)
	//		newSymbolArr = append(newSymbolArr, token.Symbol)
	//	}
	//}
	//var games []*model.Game
	//chainMap := make(map[int32][]string)
	//symbolChainMap := make(map[string]string)
	//eg.Go(func() error {
	//	games, err = model.NewGames().FindGameInfoByGameIds(gameIds) //查询游戏数据
	//	return err
	//})
	//eg.Go(func() error {
	//	chainMap, err = model.NewGameBlockChain().FindChainByGameId(gameIds) //查询游戏链数据
	//	return err
	//})
	////查找相关币价
	//eg.Go(func() error {
	//	if tokenType == "nft" {
	//		return nil
	//	}
	//	symbolChainMap, err = s.FoundCacheMarketPrice(newSymbolArr)
	//	return err
	//})
	//if err = eg.Wait(); err != nil {
	//	log.Get(s.Ctx).Errorf("FindGameInfoByAccess error(%s)", err.Error())
	//	return
	//}
	//gameMaps := make(map[int32]*model.Game)
	//for _, game := range games {
	//	gameMaps[game.ID] = game
	//}
	//if tokenType == "token" {
	//	accessBill = s.StorageToken(newTokenBt, tokenMap, gameMaps, symbolChainMap, chainMap)
	//} else if tokenType == "nft" {
	//	accessBill = s.StorageNft(newTokenBt, tokenMap, gameMaps, nftBt, chainMap)
	//}
	return
}

func (s *BalanceService) FindGameInfoByAccessV2(tokenType string, tokenBt []*TokenBt, nftBt []*NftMeta) (accessBill []*AccessBill, err error) {
	hashArr := make([]string, 0, len(tokenBt))
	newTokenBt := make([]*TokenBt, 0, len(tokenBt))
	//过滤所需要资产
	for _, token := range tokenBt {
		if tokenType == "token" {
			if token.TokenInfo.C != "" {
				continue
			}
		} else {
			if token.TokenInfo.C == "" {
				continue
			}
		}
		hashArr = append(hashArr, token.Hash)
		newTokenBt = append(newTokenBt, token)
	}
	gameIds, symbols, addressMap := model.NewGameBlockChain().FindInfoByContactAddress(hashArr) //查找资产游戏id数组
	if len(gameIds) == 0 {
		return
	}
	var games []*model.Game
	chainMap := make(map[int32][]string)
	symbolChainMap := make(map[string]string)
	tokenMap := make(map[string]*model.GameToken)
	eg := errgroup.Group{}
	eg.Go(func() error {
		games, err = model.NewGames().FindGameInfoByGameIds(gameIds) //查询游戏数据
		return err
	})
	eg.Go(func() error {
		chainMap, err = model.NewGameBlockChain().FindChainByGameId(gameIds) //查询游戏链数据
		return err
	})
	//根据游戏id查找token
	eg.Go(func() error {
		tokenMap, err = model.NewGameToken().FindTokenByGameIds(gameIds)
		return err
	})
	//查找相关币价
	eg.Go(func() error {
		if tokenType == "nft" {
			return nil
		}
		symbolChainMap, err = s.FoundCacheMarketPrice(symbols)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("FindGameInfoByAccess error(%s)", err.Error())
		return
	}
	gameMaps := make(map[int32]*model.Game)
	for _, game := range games {
		gameMaps[game.ID] = game
	}
	if tokenType == "token" {
		accessBill = s.StorageToken(newTokenBt, addressMap, gameMaps, symbolChainMap, chainMap, tokenMap)
	} else if tokenType == "nft" {
		//accessBill = s.StorageNft(newTokenBt, tokenMap, gameMaps, nftBt, chainMap)
	}
	return

}

//StorageToken 获取token资产
func (s *BalanceService) StorageToken(newTokenBt []*TokenBt, addressMap map[string]*model.GameBlockChain, gameMaps map[int32]*model.Game,
	symbolChainMap map[string]string, chainMap map[int32][]string, tokenMap map[string]*model.GameToken) (accessBill []*AccessBill) {
	gameTokenBillMap := make(map[int32][]*TokenBill)
	chainB := make(map[string]*big.Float)
	for _, bt := range newTokenBt {
		if _, ok := addressMap[bt.Hash]; !ok {
			continue
		}
		gameBlockChain := addressMap[bt.Hash]
		gameId := gameBlockChain.GameID
		if _, ok := gameMaps[gameId]; !ok {
			continue
		}
		if _, ok := tokenMap[fmt.Sprintf("%d_%s", gameId, gameBlockChain.Symbol)]; !ok {
			continue
		}
		t := tokenMap[fmt.Sprintf("%d_%s", gameId, gameBlockChain.Symbol)]
		price := symbolChainMap[fmt.Sprintf("%s_%s", gameBlockChain.Symbol, bt.NetWork)]
		fb := FormatFloat(bt.Balance, cast.ToInt(bt.TokenInfo.D))
		fba := fmt.Sprintf("%f", fb)
		fa, _ := new(big.Float).SetString(price)
		fc := fb.Mul(fb, fa)
		if v, ok := chainB[fmt.Sprintf("%d_%s", gameId, bt.NetWork)]; ok {
			v.Add(v, fc)
		} else {
			chainB[fmt.Sprintf("%d_%s", gameId, bt.NetWork)] = fc
		}
		tokenBill := &TokenBill{
			Symbol:         t.Symbol,
			Logo:           t.Logo,
			Balance:        fba,
			Price:          fmt.Sprintf("%.2f", fc),
			Chain:          bt.NetWork,
			ContactAddress: bt.Hash,
		}
		if v, ok := gameTokenBillMap[gameId]; ok {
			v = append(v, tokenBill)
			gameTokenBillMap[gameId] = v
		} else {
			tb := make([]*TokenBill, 0)
			tb = append(tb, tokenBill)
			gameTokenBillMap[gameId] = tb
		}
	}
	for gameId, bill := range gameTokenBillMap {
		chains := chainMap[gameId]
		chainBill := make([]map[string]string, 0)
		for _, chain := range chains {
			b := map[string]string{
				"chain": chain,
			}
			if v, ok := chainB[fmt.Sprintf("%d_%s", gameId, chain)]; ok {
				b["price"] = fmt.Sprintf("%.2f", v)
			} else {
				b["price"] = "0"
			}
			chainBill = append(chainBill, b)
		}
		accessBill = append(accessBill, &AccessBill{
			GameId:    gameId,
			GameName:  gameMaps[gameId].GameName,
			Logo:      gameMaps[gameId].Logo,
			Token:     bill,
			Chains:    chainMap[gameId],
			ChainBill: chainBill,
		})
	}
	return
}

//FormatFloat 设置余额字段
func FormatFloat(str string, decimals int) *big.Float {
	fbalance := new(big.Float)
	fbalance.SetString(str)
	return new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(decimals)))
}

type MarketPrice struct {
	Data   map[string][]*SymbolData `json:"data"`
	Status *MarkStatus              `json:"status"`
}

type MarkStatus struct {
	Timestamp    time.Time   `json:"timestamp"`
	ErrorCode    int         `json:"error_code"`
	ErrorMessage interface{} `json:"error_message"`
	Elapsed      int         `json:"elapsed"`
	CreditCount  int         `json:"credit_count"`
	Notice       interface{} `json:"notice"`
}

type SymbolData struct {
	Symbol   string          `json:"symbol"`
	Platform *SymbolPlatForm `json:"platform"`
	Quote    *SymbolQuote    `json:"quote"`
}

type SymbolPlatForm struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type SymbolQuote struct {
	USD map[string]interface{} `json:"USD"`
}

//FoundCacheMarketPrice 获取价格缓存,价格缓存
func (s *BalanceService) FoundCacheMarketPrice(symbolArr []string) (symbolChainMap map[string]string, err error) {
	symbols := strings.Join(symbolArr, ",")
	symbolChainMap = make(map[string]string)
	key := utils.Md5V(fmt.Sprintf("%s%s", Address_Game_Account, symbols))
	storage := CacheStorage{Data: &symbolChainMap}
	err = cache.GetStruct(key, &storage)
	if err == nil {
		if time.Now().Unix()-storage.TimeStamp > 60 {
			go s.FoundMarketPrice(key, symbolArr)
		}
		return
	}
	return s.FoundMarketPrice(key, symbolArr)
}

//FoundMarketPrice 查找市场价格
func (s *BalanceService) FoundMarketPrice(key string, symbolArr []string) (symbolChainMap map[string]string, err error) {
	symbols := strings.Join(symbolArr, ",")
	url := fmt.Sprintf("%s/cryptocurrency/quotes/latest", s.CoinMarkUrl)
	if symbols != "" {
		url = fmt.Sprintf("%s?symbol=%s", url, symbols)
	}
	headers := map[string]string{
		"X-CMC_PRO_API_KEY": s.CoinMarkApiKey,
	}
	marketRes, err := utils.NetLibGetV3(url, headers)
	if err != nil {
		log.Get(s.Ctx).Errorf("FoundMarketPrice error(%s)", err.Error())
		return
	}
	res := &MarketPrice{}
	err = json.Unmarshal(marketRes, res)
	if err != nil {
		return
	}
	if res.Status.ErrorCode != 0 {
		err = fmt.Errorf("%v", res.Status.ErrorMessage)
		return
	}
	symbolChainMap = make(map[string]string)
	dataMap := res.Data
	for _, symbol := range symbolArr {
		upSymbol := strings.ToUpper(symbol) //转为大写
		if v, ok := dataMap[upSymbol]; ok {
			for _, ts := range v {
				symbolChainMap[fmt.Sprintf("%s_%s", ts.Symbol, ts.Platform.Symbol)] = fmt.Sprintf("%v", ts.Quote.USD["price"])
			}
		}
	}
	storage := &CacheStorage{}
	storage.Data = symbolChainMap
	storage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, storage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}
