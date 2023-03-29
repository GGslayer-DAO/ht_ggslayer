package service

import (
	"encoding/json"
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/log"
	"golang.org/x/sync/errgroup"
)

type Record struct {
	Count   int         `json:"count"`
	Records interface{} `json:"records"`
}

type TokenPrice struct {
	Symbol   string   `json:"symbol"`
	Address  string   `json:"address"`
	DateList []string `json:"date_list"`
	RateList []string `json:"rate_list"`
}

type TokenHolder struct {
	Symbol   string   `json:"symbol"`
	Address  string   `json:"address"`
	DateList []string `json:"date_list"`
	NumList  []int    `json:"num_list"`
}

type TokenUserVolume struct {
	DateList   []string  `json:"date_list"`
	UsersList  []int     `json:"users_list"`
	VolumeList []float64 `json:"volume_list"`
}

//GameTokenInfo 获取游戏gameToken数据信息
func (s *GameDetailService) GameTokenInfo() (newTokenPrice []*TokenPrice, newTokenHolder []*TokenHolder,
	activeUser []int, activeVolume []float64, err error) {
	url1 := "https://www.mymetadata.io/api/v1/meta/token/price?variables="
	url2 := "https://www.mymetadata.io/api/v1/meta/token/holdGrowth?variables="
	url3 := "https://www.mymetadata.io/api/v1/meta/game/statistics?variables="
	//查询游戏数据信息
	game, err := model.NewGames().FindInfoByGameId(s.GameId, false)
	if err != nil {
		log.Get(s.Ctx).Errorf("GameTokenInfo FindInfoByGameId error(%s)", err.Error())
		return
	}
	gameName := game.GameName
	params := map[string]interface{}{
		"game_name": gameName,
		"days":      365,
	}
	var tokenPrice []*TokenPrice
	var tokenHolder []*TokenHolder
	variables := utils.MetaEncode(params)
	var tokenUserVolume *TokenUserVolume
	blockChainMap := make(map[string]string)
	//获取metadata数据
	eg := errgroup.Group{}
	eg.Go(func() error {
		url1 = fmt.Sprintf("%s%s", url1, variables)
		getRes, er := utils.NetLibGetV3(url1, nil)
		if er != nil {
			return er
		}
		record := &Record{Records: &tokenPrice}
		res := Output{Result: record}
		err = json.Unmarshal(getRes, &res)
		return err
	})
	eg.Go(func() error {
		url2 = fmt.Sprintf("%s%s", url2, variables)
		getRes, er := utils.NetLibGetV3(url2, nil)
		if er != nil {
			return er
		}
		record := &Record{Records: &tokenHolder}
		res := Output{Result: record}
		err = json.Unmarshal(getRes, &res)
		return err
	})
	eg.Go(func() error {
		url3 = fmt.Sprintf("%s%s", url3, variables)
		getRes, er := utils.NetLibGetV3(url3, nil)
		if er != nil {
			return er
		}
		res := Output{Result: &tokenUserVolume}
		err = json.Unmarshal(getRes, &res)
		return err
	})
	//根据
	eg.Go(func() error {
		chains, er := model.NewGameBlockChain().FindChainInfoByGameId(s.GameId)
		if er != nil {
			return er
		}
		for _, chain := range chains {
			blockChainMap[chain.Symbol] = chain.ContactAddress
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Get(s.Ctx).Errorf("GameTokenInfo error(%s)", err.Error())
		return
	}
	if len(tokenPrice) == 0 {
		return
	}
	tp := tokenPrice[0]
	activeUser = tokenUserVolume.UsersList
	activeVolume = tokenUserVolume.VolumeList
	for _, price := range tokenPrice {
		address := ""
		if v, ok := blockChainMap[price.Symbol]; ok {
			address = v
		}
		newTokenPrice = append(newTokenPrice, &TokenPrice{
			DateList: reverseStr(price.DateList),
			RateList: reverseStr(price.RateList),
			Symbol:   price.Symbol,
			Address:  address,
		})
	}
	for _, holder := range tokenHolder {
		dateList := make([]string, 0, len(tp.DateList))
		numList := make([]int, 0, len(tp.DateList))
		for k, t := range tp.DateList {
			dateList = append(dateList, t)
			if k < len(holder.NumList) {
				numList = append(numList, holder.NumList[k])
			} else {
				numList = append(numList, 0)
			}
		}
		address := ""
		if v, ok := blockChainMap[holder.Symbol]; ok {
			address = v
		}
		newTokenHolder = append(newTokenHolder, &TokenHolder{
			DateList: dateList,
			NumList:  numList,
			Symbol:   holder.Symbol,
			Address:  address,
		})
	}
	return
}

//reverseInt 反转int切片
func reverseInt(arr []int) []int {
	reversed := make([]int, 0, len(arr))
	//相反的顺序,并附加到新切片中
	for i := range arr {
		n := arr[len(arr)-1-i]
		reversed = append(reversed, n)
	}
	return reversed
}

func reverseStr(arr []string) []string {
	reversed := make([]string, 0, len(arr))
	//相反的顺序,并附加到新切片中
	for i := range arr {
		n := arr[len(arr)-1-i]
		reversed = append(reversed, n)
	}
	return reversed
}

//reverseFloat 反转float切片
func reverseFloat(arr []float64) []float64 {
	reversed := make([]float64, 0, len(arr))
	//相反的顺序,并附加到新切片中
	for i := range arr {
		n := arr[len(arr)-1-i]
		reversed = append(reversed, n)
	}
	return reversed
}
