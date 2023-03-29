package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/config"
	"ggslayer/utils/log"
	"golang.org/x/sync/errgroup"
	"math"
	"math/big"
	"strings"
	"time"
)

type EthResult struct {
	Id      int         `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

type EthService struct {
	Url    string
	ApiKey string
}

func NewEthService() *EthService {
	url := config.GetString("token.EthUrl")
	apiKey := config.GetString("token.EthKeyApi")
	return &EthService{
		Url:    url,
		ApiKey: apiKey,
	}
}

//GetAccountBalance 获取账户余额
func (s *EthService) GetAccountBalance(address string) *big.Float {
	url := fmt.Sprintf("%s/v2/%s", s.Url, s.ApiKey)
	params := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []string{address},
	}
	balance, err := utils.NetLibPost(url, params, nil)
	if err != nil {
		return nil
	}
	ethRes := &EthResult{}
	err = json.Unmarshal(balance, ethRes)
	if err != nil {
		return nil
	}
	ethValue := FormatFloatNumber(ethRes.Result.(string), 18)
	return ethValue
}

func FormatFloatNumber(str string, decimals int) *big.Float {
	bres := strings.ReplaceAll(str, "0x", "")
	b := utils.HexToTen(bres)
	fbalance := new(big.Float)
	fbalance.SetString(fmt.Sprintf("%d", b))
	return new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(decimals)))
}

type TokenBalance struct {
	Address       string    `json:"address"`
	TokenBalances []*TokenB `json:"tokenBalances"`
}

type TokenB struct {
	ContractAddress string         `json:"contractAddress"`
	TokenBalance    string         `json:"tokenBalance"`
	TokenMetaData   *TokenMetaData `json:"tokenMetaData"`
	TrustBalance    *big.Float     `json:"trustBalance"`
}

func (s *EthService) GetAccountToken(ty int, address string) (req *TokenBalance, err error) {
	key := utils.Md5V(fmt.Sprintf("%s%s:%d:token:", Address_Account, address, ty))
	req = new(TokenBalance)
	err = cache.GetStruct(key, &req)
	if err == nil {
		//if time.Now().Unix()-req.TimeStamp > 600 {
		//	go s.AccountNft(address, key)
		//}
		return
	}
	return s.AccountToken(address, key)
}

//AccountToken 查询代币
func (s *EthService) AccountToken(address, key string) (tr *TokenBalance, err error) {
	url := fmt.Sprintf("%s/v2/%s", s.Url, s.ApiKey)
	params := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getTokenBalances",
		"params":  []string{address},
	}
	tokenBalance, err := utils.NetLibPost(url, params, nil)
	if err != nil {
		return
	}
	tr = &TokenBalance{}
	ethRes := &EthResult{Result: tr}
	err = json.Unmarshal(tokenBalance, ethRes)
	if err != nil {
		return
	}
	eg := errgroup.Group{}
	for _, val := range tr.TokenBalances {
		newVal := val
		eg.Go(func() error {
			return s.FindTokenMetaByContractAddress(url, newVal)
		})
	}
	err = eg.Wait()
	//元数据缓存
	cache.SaveStruct(key, tr, META_CACHE_TTL) //元数据缓存时间72小时
	return
}

type TokenMetaData struct {
	Decimals int    `json:"decimals"`
	Logo     string `json:"logo"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
}

func (s *EthService) FindTokenMetaByContractAddress(url string, tokenB *TokenB) (err error) {
	params := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getTokenMetadata",
		"params":  []string{tokenB.ContractAddress},
	}
	tokenMeta, err := utils.NetLibPost(url, params, nil)
	if err != nil {
		log.New().Error(err)
		return
	}
	tm := &TokenMetaData{}
	ethRes := &EthResult{Result: tm}
	err = json.Unmarshal(tokenMeta, ethRes)
	if err != nil {
		log.New().Error(err)
		return
	}
	tokenB.TrustBalance = FormatFloatNumber(tokenB.TokenBalance, tm.Decimals)
	tokenB.TokenMetaData = tm
	return
}

type NftResult struct {
	OwnedNfts []*Nfts `json:"ownedNfts"`
	Count     int     `json:"totalCount"`
	TimeStamp int64   `json:"time_stamp"`
}

type Nfts struct {
	Balance          string            `json:"balance"`
	Contract         map[string]string `json:"contract"`
	ContractMetadata *ContractMetadata `json:"contractMetadata"`
	Media            interface{}       `json:"media"`
}

//合约元数据
type ContractMetadata struct {
	Name        string                 `json:"name"`
	OpenSea     map[string]interface{} `json:"openSea"`
	Symbol      string                 `json:"symbol"`
	TokenType   string                 `json:"tokenType"`
	TotalSupply string                 `json:"totalSupply"`
}

//GetAccountNft 查找nft元数据
func (s *EthService) GetAccountNft(ty int, address string) (req *NftResult, err error) {
	key := utils.Md5V(fmt.Sprintf("%s%s:%d:nft:", Address_Account, address, ty))
	req = new(NftResult)
	err = cache.GetStruct(key, &req)
	if err == nil {
		//if time.Now().Unix()-req.TimeStamp > 600 {
		//	go s.AccountNft(address, key)
		//}
		return
	}
	return s.AccountNft(address, key)
}

func (s *EthService) AccountNft(address, key string) (req *NftResult, err error) {
	url := fmt.Sprintf("%s/nft/v2/%s/getNFTs?owner=%s", s.Url, s.ApiKey, address)
	nftRes, err := utils.NetLibGet(url)
	if err != nil {
		log.New().Error(err)
		return
	}
	//元数据存储进redis
	req = new(NftResult)
	err = json.Unmarshal(nftRes, req)
	if err != nil {
		log.New().Error(err)
		return
	}
	req.TimeStamp = time.Now().Unix()
	//元数据缓存
	cache.SaveStruct(key, req, META_CACHE_TTL) //元数据缓存时间24小时
	return
}
