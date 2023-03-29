package validate

//游戏评论验证
type PostGameRateValidate struct {
	GameId      int32  `form:"game_id" json:"game_id" binding:"required"`
	Rate        int    `form:"rate" json:"rate" binding:"required,min=1,max=5"`
	ExperReason string `form:"exper_reason" json:"exper_reason" binding:"required"`
}

// 绑定模型获取验证错误的方法
func (r *PostGameRateValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["GameId.required"] = "game_id required"
	MsgMap["Rate.required"] = "rate required"
	MsgMap["Rate.min"] = "the rate min 1"
	MsgMap["Rate.max"] = "the rate max 5"
	MsgMap["ExperReason.required"] = "exper_reason required"
	return ForRangeValidateError(err, MsgMap)
}

//游戏新增验证
type GameSaveValidate struct {
	Name       string              `form:"name" json:"name" binding:"required"`
	GameName   string              `form:"game_name" json:"game_name" binding:"required"`
	Logo       string              `form:"logo" json:"logo" binding:"required"`
	Describe   string              `form:"describe" json:"describe"`
	Detail     string              `form:"detail" json:"detail"`
	Video      string              `form:"video" json:"video"`
	SocialLink []map[string]string `form:"social_link" json:"social_link"`
	Website    string              `form:"website" json:"website"`
	Screenshot []string            `form:"screenshot" json:"screenshot"`
	Tags       []string            `form:"tags" json:"tags"`
	Chains     []map[string]string `form:"chains" json:"chains"`
	StatusTag  string              `form:"status_tag" json:"status_tag"`
	Platform   []string            `form:"platform" json:"platform"`
}

// 绑定模型获取验证错误的方法
func (r *GameSaveValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Name.required"] = "name required"
	MsgMap["GameName.required"] = "game_name required"
	MsgMap["Logo.required"] = "logo required"
	return ForRangeValidateError(err, MsgMap)
}

//游戏验证
type GameInvestorValidate struct {
	GameId       int                 `form:"game_id" json:"game_id" binding:"required"`
	GameTerm     []map[string]string `form:"game_term" json:"game_term"`
	GameInvestor []map[string]string `form:"game_investor" json:"game_investor"`
}

// 绑定模型
func (r *GameInvestorValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["GameId.required"] = "game_id required"
	return ForRangeValidateError(err, MsgMap)
}
