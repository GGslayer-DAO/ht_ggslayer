package service

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/jwt"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"time"
)

const Jwt_User_Logout = "jwt_user_logout"

type UserService struct {
	ctx *gin.Context
}

func NewUserService(c *gin.Context) *UserService {
	return &UserService{ctx: c}
}

//服务登陆
func (s *UserService) Login(email string) (token string, err error) {
	user := &model.User{}
	//判断邮箱是否存在
	u, err := user.FindInfoByEmail(email)
	if err != nil {
		log.Get(s.ctx).Error(err)
		return
	}
	if u.ID == 0 {
		err = fmt.Errorf("email not registered, please register first")
		return
	}
	//生成jwt token
	token, err = jwt.EncretToken(uint(u.ID), false)
	err = DeleteLogout(u.ID)
	return
}

//Plogin 密码登陆
func (s *UserService) Plogin(email, password string) (token string, err error) {
	user := &model.User{}
	//判断邮箱是否存在
	u, err := user.FindInfoByEmail(email)
	if err != nil {
		log.Get(s.ctx).Error(err)
		return
	}
	if u.ID == 0 {
		err = fmt.Errorf("incorrect password")
		return
	}
	if !utils.CompareSalt(u.Password, password) {
		err = fmt.Errorf("incorrect password")
		return
	}
	token, err = jwt.EncretToken(uint(u.ID), false)
	err = DeleteLogout(u.ID)
	return
}

//Glogin 谷歌登陆
func (s *UserService) Glogin(googleEmail, name, imageUrl string) (token string, err error) {
	user := &model.User{}
	//判断邮箱是否存在
	u, err := user.FindInfoByGoogleEmail(googleEmail)
	if err != nil {
		log.Get(s.ctx).Error(err)
		return
	}
	if u.ID == 0 {
		//账号不存在，创建账号
		user.GoogleEmail = googleEmail
		user.Name = name
		user.HeadPic = imageUrl
		user.Password = utils.BcryptSalt(fmt.Sprintf("%d", time.Now().Unix()))
		user.ReferCode = CreateReferCode()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		uid, er := user.Create()
		if er != nil {
			log.Get(s.ctx).Error(err)
			return
		}
		u.ID = uid
	}
	//账号存在直接登陆
	token, err = jwt.EncretToken(uint(u.ID), false)
	err = DeleteLogout(u.ID)
	return
}

func DeleteLogout(userId int32) (err error) {
	md5key := utils.Md5V(fmt.Sprintf("%s%d", Jwt_User_Logout, userId))
	_, err = cache.RedisClient.Del(md5key).Result()
	return
}

//CreateReferCode 生成邀请码
func CreateReferCode() (referCode string) {
	scode := utils.RandNumber(12) //生成随机数字,数字缓存5分钟
	//判断邀请码数据库中是否含有，有就重新生成
	u := model.NewUser().FindReferCode(scode)
	if u.ID != 0 {
		return CreateReferCode()
	}
	return scode
}
