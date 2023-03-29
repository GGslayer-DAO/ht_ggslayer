package middleware

import (
	"fmt"
	"ggslayer/controllers/user"
	"ggslayer/utils"
	"ggslayer/utils/cache"
	"ggslayer/utils/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Auth(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		c.JSON(http.StatusUnauthorized, "token is missing")
		c.Abort()
		return
	}
	uid, err := jwt.DecretToken(authorization, false) //解密jwt
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		c.Abort()
		return
	}
	//校验黑名单是否存在
	md5key := utils.Md5V(fmt.Sprintf("%s%d", user.Jwt_User_Logout, uid))
	suid, _ := cache.RedisClient.Get(md5key).Result()
	if suid != "" {
		c.JSON(http.StatusUnauthorized, "Token is expired")
		c.Abort()
		return
	}
	c.Set("user_id", uid)
}

func MustNotAuth(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	if authorization != "" {
		uid, err := jwt.DecretToken(authorization, false) //解密jwt
		if err != nil {
			c.Next()
			return
		}
		c.Set("user_id", uid)
	}
}

func AdmAuth(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		c.JSON(http.StatusUnauthorized, "token is missing")
		c.Abort()
		return
	}
	uid, err := jwt.DecretToken(authorization, true) //解密jwt
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		c.Abort()
		return
	}
	//校验黑名单是否存在
	md5key := utils.Md5V(fmt.Sprintf("%s%d", user.Jwt_User_Logout, uid))
	suid, _ := cache.RedisClient.Get(md5key).Result()
	if suid != "" {
		c.JSON(http.StatusUnauthorized, "Token is expired")
		c.Abort()
		return
	}
	c.Set("user_id", uid)
}
