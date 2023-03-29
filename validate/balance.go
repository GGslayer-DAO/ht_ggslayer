package validate

//绑定地址校验
type BindTokenValidate struct {
	TokenType int    `form:"token_type" json:"token_type" binding:"required"`
	Address   string `form:"address" json:"address" binding:"required"`
}

func (r *BindTokenValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["TokenType.required"] = "token_type required"
	MsgMap["Address.required"] = "headpic required"
	return ForRangeValidateError(err, MsgMap)
}
