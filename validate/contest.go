package validate

//比赛新增验证
type ContestSaveValidate struct {
	Title     string  `form:"title" json:"title" binding:"required"`
	StartTime string  `form:"start_time" json:"start_time" binding:"required"`
	EndTime   string  `form:"end_time" json:"end_time" binding:"required"`
	GameArr   []int32 `form:"game_arr" json:"game_arr" binding:"required"`
}

// 绑定模型获取验证错误的方法
func (r *ContestSaveValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Title.required"] = "title required"
	MsgMap["StartTime.required"] = "start_time required"
	MsgMap["EndTime.required"] = "end_time required"
	MsgMap["GameArr.required"] = "game_arr required"
	return ForRangeValidateError(err, MsgMap)
}
