package game

import (
	"ggslayer/utils"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
)

type DapResult struct {
	Code int      `json:"code"`
	Data *DapData `json:"data"`
}

type DapData struct {
	TotalPages int       `json:"totalPages"`
	TotalRows  int       `json:"totalRows"`
	DapResult  []*DapRes `json:"result"`
}

type DapRes struct {
	BgLogo       string            `json:"bgLogo"`
	CurrentPrice string            `json:"currentPrice"`
	Logo         string            `json:"logo"`
	Name         string            `json:"name"`
	PlatForm     string            `json:"platform"`
	PlatForms    map[string]string `json:"platforms"`
	Symbol       string            `json:"symbol"`
}

func DapGameCraw(c *gin.Context) {
	url := "https://app-api-prod.dapdap.io/api/game/query.json"
	params := map[string]interface{}{
		"orderBy":  1,
		"pageNum":  1,
		"pageSize": 50,
	}
	headers := map[string]interface{}{
		"content-type": "application/json",
	}
	res, _ := utils.NetLibPost(url, params, headers)

	log.PP(string(res))
}
