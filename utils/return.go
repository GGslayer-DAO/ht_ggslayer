package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	// 代码
	Code int `json:"code" example:"200"`
	// 数据集
	Data interface{} `json:"data"`
	// 消息
	Msg string `json:"message"`
}

// 失败数据处理
func Error(c *gin.Context, code int, errMsg string) {
	var res Response
	res.Code = code
	res.Msg = errMsg
	c.JSON(http.StatusOK, res)
}

// 通常成功数据处理
func Success(c *gin.Context, data interface{}, msg string) {
	var res Response
	res.Data = data
	if msg != "" {
		res.Msg = msg
	}
	res.Code = 200
	c.JSON(http.StatusOK, res)
}

func PageReturn(page, size, total, lastPage int, list interface{}) map[string]interface{} {
	return map[string]interface{}{
		"current_page": page,
		"next_page":    page + 1,
		"total":        total,
		"list":         list,
		"last_page":    lastPage,
		"size":         size,
	}
}
