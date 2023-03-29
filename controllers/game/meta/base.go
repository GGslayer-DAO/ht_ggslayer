package meta

import (
	"encoding/json"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Output struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Result  *Result `json:"result"`
}

type Result struct {
	Count   int        `json:"count"`
	Records []*Records `json:"records"`
}

type Records struct {
	GameName   string            `json:"game_name"`
	Name       string            `json:"name"`
	Logo       string            `json:"logo"`
	BlockChain []string          `json:"blockchain"`
	Tags       []string          `json:"tags"`
	Balance    map[string]string `json:"balance"`
	ActiveUser *ActiveUser       `json:"active_user"`
	Social     *Social           `json:"social"`
	Holders    *Holders          `json:"holders"`
}

type ActiveUser struct {
	Rate  string `json:"rage"`
	Value int    `json:"value"`
}

type Social struct {
	Rate  string `json:"rage"`
	Value int    `json:"value"`
}

type Holders struct {
	Rate  string `json:"rage"`
	Value int    `json:"value"`
}

type OutputToken struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Result  [][]*Token `json:"result"`
}

type Token struct {
	Symbol        string `json:"symbol"`
	Price         string `json:"price"`
	Logo          string `json:"logo"`
	ChangeRate24H string `json:"change_rate_24h"`
	Holders       int    `json:"holders"`
}

type ChainRecords struct {
	Name     string `json:"name"`
	ShowName string `json:"show_name"`
}

func BaseCraw(c *gin.Context, index int) (output *Output, outputToken *OutputToken, cmap map[string]string, err error) {
	//爬取游戏数据
	variables := ArrGames[index]
	url := "https://www.mymetadata.io/api/v1/meta/game/list?variables=" + variables
	output = new(Output)
	outputToken = new(OutputToken)
	eg := errgroup.Group{}
	//爬取游戏数据
	eg.Go(func() error {
		res, er := utils.NetLibGet(url)
		if er != nil {
			return er
		}
		err = json.Unmarshal(res, output)
		return err
	})
	//爬取游戏token数据
	urlPath := "https://www.mymetadata.io/api/v1/meta/game/list/token"
	params := map[string]string{
		"variables": ArrTokens[index],
	}
	eg.Go(func() error {
		resT, er := utils.NetPostForm(urlPath, params)
		if er != nil {
			return er
		}
		err = json.Unmarshal(resT, outputToken)
		return err
	})
	cmap = make(map[string]string)
	//获取短链数据信息
	eg.Go(func() error {
		key := utils.Md5V("chain_list")
		var chainRecord []*ChainRecords
		cache.GetStruct(key, &chainRecord)
		for _, sage := range chainRecord {
			cmap[sage.Name] = sage.ShowName
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Get(c).Errorf("MetaGameCraw error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
	}
	return
}
