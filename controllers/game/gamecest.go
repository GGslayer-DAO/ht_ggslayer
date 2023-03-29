package game

import (
	"ggslayer/tasks/meta"
	"github.com/gin-gonic/gin"
)

func Cest(c *gin.Context) {
	//params := map[string]interface{}{
	//	"game_name": "alien-worlds",
	//}

	//curl --header "X-API-KEY: [YOUR_API_KEY]" --request GET -i --url 'https://api.opensea.io/api/v1/assets'
	//s := utils.MetaEncode(params)
	//str := "eWZpNm45ZXlKbllXMWxYMjVoYldVaU9pSmhiR2xsYmkxM2IzSnNaSE1pTENKa1lYbHpJam96TmpWOWMzOWx6cGs%3D"
	//log.PP(utils.MetaDecode(str))
	//meta.GameChainDetail()
	meta.GameDetail()
}
