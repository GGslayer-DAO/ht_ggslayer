package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/config"
	"ggslayer/utils/log"
	"strings"
	"time"
)

const (
	ImageUrl = "https://gateway.pinata.cloud/ipfs/"
	GameNft  = "game_nft:"
)

var chainMap = map[string]int{
	"eth":       1,
	"polygon":   137,
	"bsc":       56,
	"avalanche": 43114,
	"fantom":    250,
}

type NftInfo struct {
	ContractAddress string    `json:"contract_address"`
	ErcType         string    `json:"erc_type"`
	ImageUri        string    `json:"image_uri"`
	Name            string    `json:"name"`
	Owner           string    `json:"owner"`
	Symbol          string    `json:"symbol"`
	TokenId         string    `json:"token_id"`
	TokenUri        string    `json:"token_uri"`
	Total           int       `json:"total"`
	Chain           string    `json:"chain"`
	TotalString     string    `json:"total_string"`
	FloorPrice      float64   `json:"floor_price"`
	Metadata        *Metadata `json:"metadata"`
}

type Metadata struct {
	AnimationUrl string `json:"animation_url"`
	Description  string `json:"description"`
	ExternalUrl  string `json:"external_url"`
	Image        string `json:"image"`
}

//GetCacheGameNft 获取缓存nft数据
func (s *GameDetailService) GetCacheGameNft() (nftInfos []*NftInfo, err error) {
	key5v := utils.Md5V(fmt.Sprintf("%s%d", "game_nft", s.GameId))
	key := fmt.Sprintf("%s%s", GameNft, key5v)
	storage := &CacheStorage{
		Data: &nftInfos,
	}
	err = cache.GetStruct(key, &storage)
	if err == nil {
		if time.Now().Unix()-storage.TimeStamp > 600 {
			go s.FindGameNft(key)
		}
		return
	}
	return s.FindGameNft(key)
}

//FindGameNft 查找游戏nft数据
func (s *GameDetailService) FindGameNft(key string) (nftInfos []*NftInfo, err error) {
	chains, err := model.NewGameBlockChain().FindChainInfoByGameId(s.GameId)
	if err != nil {
		log.Get(s.Ctx).Errorf("FindGameNft FindChainInfoByGameId error(%s)", err.Error())
		return
	}
	if len(chains) == 0 {
		return
	}
	apikey := config.GetString("token.ChainBaseApiKey")
	for _, chain := range chains {
		address := chain.ContactAddress
		chainId := 0
		if address == "" {
			continue
		}
		shortName := strings.ToLower(chain.ShortName)
		if v, ok := chainMap[shortName]; ok {
			chainId = v
		} else {
			continue
		}
		url := "https://api.chainbase.online/v1/account/nfts"
		url = fmt.Sprintf("%s?address=%s&chain_id=%d", url, address, chainId)
		headers := map[string]string{
			"X-API-KEY": apikey,
		}
		getRes, er := utils.NetLibGetV3(url, headers)
		if er != nil {
			log.Get(s.Ctx).Errorf("error(%s)", er.Error())
			return
		}
		res := TokenResult{Data: &nftInfos}
		err = json.Unmarshal(getRes, &res)
		if err != nil {
			return
		}
		if len(nftInfos) == 0 {
			continue
		}
		for _, nft := range nftInfos {
			//计算地板价
			nft.FloorPrice, _ = s.CalFloorPrice(nft.ContractAddress, apikey, chainId)
			//图片转化
			nft.ImageUri = strings.ReplaceAll(nft.ImageUri, "ipfs://", ImageUrl)
			//tokenId转化
			bres := strings.ReplaceAll(nft.TokenId, "0x", "")
			b := utils.HexToTen(bres)
			nft.TokenId = fmt.Sprintf("%d", b)
			nft.Chain = chain.ShortName
		}
	}
	if len(nftInfos) == 0 {
		return
	}
	storage := &CacheStorage{}
	storage.Data = nftInfos
	storage.TimeStamp = time.Now().Unix()
	err = cache.SaveStruct(key, storage, META_CACHE_TTL) //元数据缓存时间72小时
	return
}

type FloorSt struct {
	FloorPrice float64   `json:"floor_price"`
	Symbol     string    `json:"symbol"`
	Source     string    `json:"source"`
	UpdatedAt  time.Time `json:"updated_at"`
}

//CalFloorPrice 计算地板价
func (s *GameDetailService) CalFloorPrice(contractAddress, apikey string, chainId int) (floorPrice float64, err error) {
	url := "https://api.chainbase.online/v1/nft/floor_price"
	url = fmt.Sprintf("%s?contract_address=%s&chain_id=%d", url, contractAddress, chainId)
	headers := map[string]string{
		"X-API-KEY": apikey,
	}
	getRes, er := utils.NetLibGetV3(url, headers)
	if er != nil {
		log.Get(s.Ctx).Errorf("error(%s)", er.Error())
		return
	}
	var floorst *FloorSt
	res := TokenResult{Data: &floorst}
	err = json.Unmarshal(getRes, &res)
	if err != nil {
		return
	}
	if res.Code != 0 {
		return
	}
	floorPrice = floorst.FloorPrice
	return
}
