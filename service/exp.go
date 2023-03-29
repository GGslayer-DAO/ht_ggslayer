package service

import "ggslayer/model"

//AddExp 增加经验值
func AddExp(userId, etype, exp int) {
	go model.NewExp().CreateExp(userId, etype, exp)
}

//GameCollectAddExp 游戏收藏增加经验值
func GameCollectAddExp(userId, etype, exp int, isAdd bool) {
	go func() {
		collects := model.NewUserGameCollect().FindByUserId(userId)
		if len(collects) >= 2 || len(collects) == 0 {
			return
		}
		if !isAdd {
			return
		}
		model.NewExp().CreateExp(userId, etype, exp)
	}()
}

//GuideCollectAddExp 攻略收藏增加经验值
func GuideCollectAddExp(userId, etype, exp int, isAdd bool) {
	go func() {
		collects := model.NewUserGuideCollect().FindByUserId(userId)
		if len(collects) >= 2 || len(collects) == 0 {
			return
		}
		if !isAdd {
			return
		}
		model.NewExp().CreateExp(userId, etype, exp)
	}()
}
