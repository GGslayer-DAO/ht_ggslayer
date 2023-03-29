package validate

//活动新增验证
type EventSaveValidate struct {
	Name          string                   `form:"name" json:"name" binding:"required"`
	Describe      string                   `form:"describe" json:"describe" binding:"required"`
	Image         string                   `form:"image" json:"image" binding:"required"`
	Url           string                   `form:"url" json:"url"`
	Type          string                   `form:"type" json:"type" binding:"required,oneof=IGO INO Airdrop"`
	Tags          []string                 `form:"tags" json:"tags"`
	EndTime       string                   `form:"end_time" json:"end_time" binding:"required"`
	TotalAward    float64                  `form:"total_award" json:"total_award"`
	Price         float64                  `form:"price" json:"price"`
	TotalCount    int                      `form:"total_count" json:"total_count"`
	Chain         []map[string]interface{} `form:"chain" json:"chain"`
	AirdropReward string                   `form:"airdrop_reward" json:"airdrop_reward"`
	Quota         int                      `form:"quota" json:"quota"`
}

// 绑定模型获取验证错误的方法
func (r *EventSaveValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Name.required"] = "name required"
	MsgMap["Describe.required"] = "describe required"
	MsgMap["Image.required"] = "image required"
	MsgMap["Type.required"] = "type required"
	MsgMap["Type.oneof"] = "type must one of IGO,INO or Airdrop"
	MsgMap["EndTime.required"] = "end_time required"
	return ForRangeValidateError(err, MsgMap)
}
