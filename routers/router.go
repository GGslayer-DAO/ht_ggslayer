package routers

import (
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := &gin.Engine{}
	r = gin.New()

	r.Use(log.GinLogger()) // 记录日志
	api(r)
	adm(r)
	validate.CustomFunValidate() //添加自定义验证方法
	return r
}
