package user

import (
	"fmt"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/config"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"time"
)

const Jwt_User_Logout = "jwt_user_logout"

//code login
func CodeLogin(c *gin.Context) {
	var v validate.CodeLoginValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//验证code是否过期
	code := v.Code
	b := CheckCode(v.Email, "login", code)
	if !b {
		ginc.Fail(c, "incorrect verification code", "400")
		return
	}
	s := service.NewUserService(c)
	token, err := s.Login(v.Email)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, token, "success")
}

//PassLogin login
func PassLogin(c *gin.Context) {
	var v validate.PassLoginValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}

	s := service.NewUserService(c)
	token, err := s.Plogin(v.Email, v.Password)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, token, "success")
}

//Logout 用户登出
func Logout(c *gin.Context) {
	userId := c.GetInt("user_id")
	md5key := utils.Md5V(fmt.Sprintf("%s%d", Jwt_User_Logout, userId))
	ttl := config.GetInt("jwt.Ttl")
	_, err := cache.RedisClient.Set(md5key, 1, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error())
		return
	}
	ginc.Ok(c, "success")
}

//GoogleLogin 谷歌登陆绑定用户信息
func GoogleLogin(c *gin.Context) {
	var v validate.GoogleLoginValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Errorf("GoogleLogin error(%s)", err.Error())
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	s := service.NewUserService(c)
	token, err := s.Glogin(v.Email, v.Name, v.ImageUrl)
	if err != nil {
		log.Get(c).Errorf("GoogleLogin error(%s)", err.Error())
		ginc.Fail(c, "google login fail", "400")
		return
	}
	ginc.Ok(c, token, "success")
}
