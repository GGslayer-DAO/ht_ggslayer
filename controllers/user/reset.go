package user

import (
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
)

//ResetFirst 忘记密码邮箱确认
func ResetFirst(c *gin.Context) {
	var v validate.ResetFirstValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	user, err := model.NewUser().FindInfoByEmail(v.Email)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if user.ID == 0 {
		ginc.Fail(c, "mailbox does not exist", "400")
		return
	}
	ginc.Ok(c, "success")
}

//ResetPass 忘记密码重置密码
func ResetPass(c *gin.Context) {
	var v validate.ResetPassValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//验证code是否过期
	code := v.Code
	b := CheckCode(v.Email, "reset", code)
	if !b {
		ginc.Fail(c, "incorrect verification code", "400")
		return
	}
	//判断邮箱是否存在
	u, err := model.NewUser().FindInfoByEmail(v.Email)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if u.ID == 0 {
		ginc.Fail(c, "mailbox does not exist")
		return
	}
	u.Password = utils.BcryptSalt(v.Password)
	err = u.Update()
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//ToResetPass 用户重置密码
func ToResetPass(c *gin.Context) {
	var v validate.ToResetPassValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	userId := c.GetInt("user_id")
	user, err := model.NewUser().FindById(userId)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "reset password fail", "400")
		return
	}
	//校验原始密码是否正确
	if !utils.CompareSalt(user.Password, v.OriginalPassword) {
		ginc.Fail(c, "incorrect original_password", "400")
		return
	}
	user.Password = utils.BcryptSalt(v.Password)
	err = user.Update()
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "reset password fail", "400")
		return
	}
	ginc.Ok(c, "success")
}
