package validate

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

type Validate interface {
	GetError(err validator.ValidationErrors) string
}

/**
 * 错误数据返回
 */
func TranslateValidateError(err error, msgMap map[string]string) map[string]string {
	var errMap = map[string]string{}
	if err != nil {
		logrus.Errorf(err.Error())
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			key := fmt.Sprintf("%v.%v", e.Field(), e.Tag())
			if _, ok := msgMap[key]; ok {
				if e.Param() != "" {
					errMap[key] = strings.Replace(msgMap[key], "{"+e.Tag()+"}", e.Param(), -1)
				} else {
					errMap[key] = msgMap[key]
				}
			} else {
				errMap[key] = key + "未定义翻译字段"
			}
		}
	}
	return errMap
}

/**
 * 循环验证错误
 */
func ForRangeValidateError(err error, MsgMap map[string]string) string {
	errorMap := TranslateValidateError(err, MsgMap)
	for _, value := range errorMap {
		return value
	}
	return "参数错误"
}

func CustomFunValidate() {
	//将验证方法注册到验证器中
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("phone", phone)
		v.RegisterValidation("charNumber", charNumer)
	}
}

//手机自定义验证
func phone(c validator.FieldLevel) bool {
	b, _ := regexp.MatchString(`^1([38][0-9]|4[579]|5[0-3,5-9]|6[6]|7[0135678]|9[189])\d{8}$`, c.Field().Interface().(string))
	return b
}

//验证含有数字和大小写字母
func charNumer(c validator.FieldLevel) bool {
	b, _ := regexp.MatchString(`[A-Z][a-z]?\d*`, c.Field().Interface().(string))
	return b
}
