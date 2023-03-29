package user

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"time"
)

const EmailCache = "email_cache:"

//PostEmail 发送邮件
func PostEmail(c *gin.Context) {
	var v validate.PostEmailValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	md5lock := utils.Md5V(fmt.Sprintf("%s%s", v.Email, v.Type))
	//控制发送频次，60s内不能重新发送
	b, err := cache.RedisClient.SetNX(fmt.Sprintf("%s%s", md5lock, "lock"), 1, 60*time.Second).Result()
	if err != nil {
		log.Get(c).Errorf(err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if !b {
		ginc.Fail(c, "frequently submit", "400")
		return
	}
	sn := utils.RandNumber(6) //生成随机数字,数字缓存5分钟
	md5sign := utils.Md5V(fmt.Sprintf("%s%s%s", v.Email, v.Type, sn))
	//添加缓存code
	_, err = cache.RedisClient.Set(fmt.Sprintf("%s%s", EmailCache, md5sign), sn, 300*time.Second).Result()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	go CheckSend(v.Email, fmt.Sprintf("ggslayer %s code", v.Type), sn, v.Type)
	ginc.Ok(c, "success")
}

//CheckSend 检验发送
func CheckSend(mailTo, subject, body, ty string) {
	mail := utils.NewMail()
	mail.SetMailInfo(mailTo, subject, body)
	err := mail.Send()
	if err != nil {
		log.New().Errorf(ty, err.Error())
	}
}

//CheckCode 缓存验证码校验
func CheckCode(email, t, code string) bool {
	md5sign := utils.Md5V(fmt.Sprintf("%s%s%s", email, t, code))
	scode, err := cache.RedisClient.Get(fmt.Sprintf("%s%s", EmailCache, md5sign)).Result()
	if err != nil {
		return false
	}
	if scode != code {
		return false
	}
	return true
}

//ChangeEmail 修改邮箱
func ChangeEmail(c *gin.Context) {
	var v validate.ChangeEmailValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	userId := c.GetInt("user_id")
	//验证code是否过期
	code := v.Code
	b := CheckCode(v.Email, "change_email", code)
	if !b {
		ginc.Fail(c, "incorrect verification code", "400")
		return
	}
	user, err := model.NewUser().FindById(userId)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "change email fail", "400")
		return
	}
	//校验密码是否正确
	if !utils.CompareSalt(user.Password, v.Password) {
		ginc.Fail(c, "incorrect password", "400")
		return
	}
	//判断邮箱是否存在
	u, err := model.NewUser().FindInfoByEmail(v.Email)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "change email fail", "400")
		return
	}
	if u.ID != 0 {
		ginc.Fail(c, "mailbox already exists", "400")
		return
	}
	user.Email = v.Email
	err = user.Update()
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "change email fail", "400")
		return
	}
	ginc.Ok(c, "success")
}

//BindEmail 谷歌登陆绑定邮箱
func BindEmail(c *gin.Context) {
	
}
