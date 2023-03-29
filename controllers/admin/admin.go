package admin

import (
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/jwt"
	"github.com/gin-gonic/gin"
)

type Adm struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

//Create 创建管理员
func Create(c *gin.Context) {
	var v Adm
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	admins := &model.Admin{}
	admins.UserName = v.UserName
	admins.Password = utils.BcryptSalt(v.Password)
	admins.Create()
	ginc.Ok(c, "success")
}

//管理员登陆
func Login(c *gin.Context) {
	var v Adm
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if v.UserName == "" {
		ginc.Fail(c, "please input username", "400")
		return
	}
	if v.Password == "" {
		ginc.Fail(c, "please input password", "400")
		return
	}
	admins, err := model.NewAdmin().FindByUserName(v.UserName)
	if err != nil || admins.ID == 0 {
		ginc.Fail(c, "incorrect username or password", "400")
		return
	}
	if !utils.CompareSalt(admins.Password, v.Password) {
		ginc.Fail(c, "incorrect username or password", "400")
		return
	}
	//生成jwt token
	token, err := jwt.EncretToken(uint(admins.ID), true)
	if err != nil {
		ginc.Fail(c, "incorrect username or password", "400")
	}
	ginc.Ok(c, token)
}
