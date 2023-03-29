package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/config"
	"ggslayer/utils/log"
	"golang.org/x/sync/errgroup"
	"math/big"
	"strings"
	"sync"
	"time"
)

const UserBalance = "user_balance:"

type TokenResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ToB struct {
	Balance         string `json:"balance"`
	ContractAddress string `json:"contract_address"`
	Decimals        int    `json:"decimals"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
}

func (s *BalanceService) GetCacheUserAddressBalance() (allTob []*ToB, nftInfos []*NftInfo, err error) {
	key5v := utils.Md5V(fmt.Sprintf("%s%s", "user_balance", s.Address))
	key := fmt.Sprintf("%s%s", UserBalance, key5v)
	storage := &AccessStorage{
		Token: &allTob,
		Nft:   &nftInfos,
	}
	err = cache.GetStruct(key, &storage)
	if err == nil {
		if time.Now().Unix()-storage.TimeStamp > 600 {
			go s.FindUserAddressBalance(key)
		}
		return
	}
	return s.FindUserAddressBalance(key)
}

//FindUserAddressBalance 查询用户地址资产
func (s *BalanceService) FindUserAddressBalance(key string) (allTob []*ToB, nftInfos []*NftInfo, err error) {
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
	lock1 := sync.Mutex{}
	lock2 := sync.Mutex{}
	headers := map[string]string{
		"X-API-KEY": config.GetString("token.ChainBaseApiKey"),
	}
	eg := errgroup.Group{}
	for _, addrB := range addressB {
		newAddrB := addrB
		network := strings.ToLower(newAddrB.Network)
		chainId := 0
		if _, ok := chainMap[network]; !ok {
			continue
		}
		chainId = chainMap[network]
		//查询代币数据信息
		eg.Go(func() error {
			turl := fmt.Sprintf("https://api.chainbase.online/v1/account/tokens?address=%s&chain_id=%d&limit=100", s.Address, chainId)
			tokenRes, er := utils.NetLibGetV3(turl, headers)
			if er != nil {
				return er
			}
			var tob []*ToB
			rs := TokenResult{Data: &tob}
			err = json.Unmarshal(tokenRes, &rs)
			if err != nil {
				return err
			}
			lock1.Lock()
			allTob = append(allTob, tob...)
			lock1.Unlock()
			return nil
		})
		//查找nft数据信息
		eg.Go(func() error {
			nurl := fmt.Sprintf("https://api.chainbase.online/v1/account/nfts?address=%s&chain_id=%d&limit=100", s.Address, chainId)
			nftRes, er := utils.NetLibGetV3(nurl, headers)
			if er != nil {
				return er
			}
			var nfts []*NftInfo
			rs := TokenResult{Data: &nfts}
			err = json.Unmarshal(nftRes, &rs)
			if err != nil {
				return err
			}
			lock2.Lock()
			nftInfos = append(nftInfos, nfts...)
			lock2.Unlock()
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("FindAddressBalance error(%s)", err.Error())
	}
	accessStorage := &AccessStorage{}
	accessStorage.Token = allTob
	accessStorage.Nft = nftInfos
	accessStorage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, accessStorage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}

//FindGameInfoByTob 根据代币查找游戏资产
func (s *BalanceService) FindGameInfoByTob(tobs []*ToB) (accessBill []*AccessBill, err error) {
	contactArr := make([]string, 0, len(tobs)) //查找代币合约地址
	for _, tob := range tobs {
		contactArr = append(contactArr, tob.ContractAddress)
	}
	gameIds, symbols, addressMap := model.NewGameBlockChain().FindInfoByContactAddress(contactArr) //查找资产游戏id数组
	if len(gameIds) == 0 {
		return
	}
	var games []*model.Game
	gameChainMap := make(map[int32][]string)
	symbolChainMap := make(map[string]string)
	tokenMap := make(map[string]*model.GameToken)
	eg := errgroup.Group{}
	eg.Go(func() error {
		games, err = model.NewGames().FindGameInfoByGameIds(gameIds) //查询游戏数据
		return err
	})
	eg.Go(func() error {
		gameChainMap, err = model.NewGameBlockChain().FindChainByGameId(gameIds) //查询游戏链数据
		return err
	})
	//根据游戏id查找token
	eg.Go(func() error {
		tokenMap, err = model.NewGameToken().FindTokenByGameIds(gameIds)
		return err
	})
	//查找相关币价
	eg.Go(func() error {
		symbolChainMap, err = s.FoundCacheMarketPrice(symbols)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("FindGameInfoByTob error(%s)", err.Error())
		return
	}
	gameMaps := make(map[int32]*model.Game)
	for _, game := range games {
		gameMaps[game.ID] = game
	}
	accessBill = s.StorageUserToken(tobs, addressMap, gameMaps, symbolChainMap, gameChainMap, tokenMap)
	return
}

//StorageUserToken 获取token资产
func (s *BalanceService) StorageUserToken(tobs []*ToB, addressMap map[string]*model.GameBlockChain, gameMaps map[int32]*model.Game,
	symbolChainMap map[string]string, chainMap map[int32][]string, tokenMap map[string]*model.GameToken) (accessBill []*AccessBill) {
	gameTokenBillMap := make(map[int32][]*TokenBill)
	chainB := make(map[string]*big.Float)
	for _, tb := range tobs {
		if _, ok := addressMap[tb.ContractAddress]; !ok {
			continue
		}
		gameBlockChain := addressMap[tb.ContractAddress]
		gameId := gameBlockChain.GameID
		if _, ok := gameMaps[gameId]; !ok {
			continue
		}
		if _, ok := tokenMap[fmt.Sprintf("%d_%s", gameId, gameBlockChain.Symbol)]; !ok {
			continue
		}
		t := tokenMap[fmt.Sprintf("%d_%s", gameId, gameBlockChain.Symbol)]
		price := symbolChainMap[fmt.Sprintf("%s_%s", gameBlockChain.Symbol, gameBlockChain.ShortName)]
		bres := strings.ReplaceAll(tb.Balance, "0x", "")
		fb := FormatFloat(utils.HexToTen(bres).String(), tb.Decimals)
		fba := fmt.Sprintf("%f", fb)
		fa, _ := new(big.Float).SetString(price)
		fc := fb.Mul(fb, fa)
		if v, ok := chainB[fmt.Sprintf("%d_%s", gameId, gameBlockChain.ShortName)]; ok {
			v.Add(v, fc)
		} else {
			chainB[fmt.Sprintf("%d_%s", gameId, gameBlockChain.ShortName)] = fc
		}
		tokenBill := &TokenBill{
			Symbol:         t.Symbol,
			Logo:           t.Logo,
			Balance:        fba,
			Price:          fmt.Sprintf("%.2f", fc),
			Chain:          gameBlockChain.ShortName,
			ContactAddress: tb.ContractAddress,
		}
		if v, ok := gameTokenBillMap[gameId]; ok {
			v = append(v, tokenBill)
			gameTokenBillMap[gameId] = v
		} else {
			gameTokenBillMap[gameId] = []*TokenBill{tokenBill}
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

//FindGameInfoByNft
func (s *BalanceService) FindGameInfoByNft(nfts []*NftInfo) (accessBill []*AccessBill, err error) {
	nftMap := make(map[string]interface{})
	nftArr := make([]string, 0, len(nfts))
	for _, nft := range nfts {
		if _, ok := nftMap[nft.ContractAddress]; !ok {
			nftMap[nft.ContractAddress] = 1
			nftArr = append(nftArr, nft.ContractAddress)
		}
	}
	gameIds, _, addressMap := model.NewGameBlockChain().FindInfoByContactAddress(nftArr) //查找资产游戏id数组
	if len(gameIds) == 0 {
		return
	}
	var games []*model.Game
	eg := errgroup.Group{}
	eg.Go(func() error {
		games, err = model.NewGames().FindGameInfoByGameIds(gameIds) //查询游戏数据
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("FindGameInfoByNft error(%s)", err.Error())
		return
	}
	gameMaps := make(map[int32]*model.Game)
	for _, game := range games {
		gameMaps[game.ID] = game
	}
	accessBill = s.StorageUserNft(nfts, addressMap, gameMaps)
	return
}

//StorageUserNft 获取nft资产
func (s *BalanceService) StorageUserNft(nfts []*NftInfo, addressMap map[string]*model.GameBlockChain, gameMaps map[int32]*model.Game) (accessBill []*AccessBill) {
	gameNftBillMap := make(map[int32][]*NftInfo)
	for _, nft := range nfts {
		var gameId int32
		if _, ok := addressMap[nft.ContractAddress]; !ok {
			continue
		}
		blockChain := addressMap[nft.ContractAddress]
		gameId = blockChain.GameID
		if _, ok := gameMaps[gameId]; !ok {
			continue
		}
		if v, ok := gameNftBillMap[gameId]; ok {
			v = append(v, nft)
			gameNftBillMap[gameId] = v
		} else {
			gameNftBillMap[gameId] = []*NftInfo{nft}
		}
	}
	for gameId, bill := range gameNftBillMap {
		accessBill = append(accessBill, &AccessBill{
			GameId:   gameId,
			GameName: gameMaps[gameId].GameName,
			Logo:     gameMaps[gameId].Logo,
			Nft:      bill,
		})
	}
	return
}
