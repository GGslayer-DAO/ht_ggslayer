package validate

//攻略新增验证
type GuideSaveValidate struct {
	GameId      int32  `form:"game_id" json:"game_id" binding:"required"`
	Title       string `form:"title" json:"title" binding:"required"`
	Describe    string `form:"describe" json:"describe"`
	Content     string `form:"content" json:"content"`
	Url         string `form:"url" json:"url"`
	Author      string `form:"author" json:"author"`
	AuthorImage string `form:"author_image" json:"author_image"`
}

// 绑定模型获取验证错误的方法
func (r *GuideSaveValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["GameId.required"] = "game_id required"
	MsgMap["Title.required"] = "title required"
	MsgMap["Describe.required"] = "describe required"
	return ForRangeValidateError(err, MsgMap)
}
