package user

import (
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/jwt"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"time"
)

//register
func Register(c *gin.Context) {
	var v validate.RegisterValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//验证code是否过期
	code := v.Code
	b := CheckCode(v.Email, "register", code)
	if !b {
		ginc.Fail(c, "incorrect verification code", "400")
		return
	}
	user := &model.User{}
	//判断邮箱是否存在
	u, err := user.FindInfoByEmail(v.Email)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if u.ID != 0 {
		ginc.Fail(c, "email already exists, please log in", "400")
		return
	}
	//创建账号
	user.Email = v.Email
	user.Password = utils.BcryptSalt(v.Password)
	user.ReferCode = service.CreateReferCode()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	uid, err := user.Create()
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "user create error", "400")
		return
	}
	//注册后直接登陆
	token, err := jwt.EncretToken(uint(uid), false)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "jwt encret token error", "400")
		return
	}
	//注册成功，获得10经验值
	service.AddExp(int(uid), 1, 10)
	//todo 有邀请码，处理邀请码相关
	ginc.Ok(c, token)
}
