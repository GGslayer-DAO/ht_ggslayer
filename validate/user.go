package validate


//用户信息更新校验
type UserUpdateValidate struct {
	Name    string `form:"name" json:"name" binding:"required"`
	HeadPic string `form:"headpic" json:"headpic" binding:"required"`
	Describe string `form:"describe" json:"describe" binding:"required"`
}

func (r *UserUpdateValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Name.required"] = "display_name required"
	MsgMap["HeadPic.required"] = "headpic required"
	MsgMap["Describe.required"] = "describe required"
	return ForRangeValidateError(err, MsgMap)
}
