package admin

import (
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"math"
	"time"
)

type GuideInfo struct {
	Id          int32  `json:"id"`
	GameId      int32  `json:"game_id"`
	Title       string `json:"title"`        // 标题
	Describe    string `json:"describe"`     // 简介
	URL         string `json:"url"`          // 图片链接封面
	Author      string `json:"author"`       // 作者
	AuthorImage string `json:"author_image"` // 作者头像
	GameName    string `json:"game_name"`    //游戏名称
	Status      int    `json:"status"`
}

//GuideList 攻略列表
func GuideList(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	status := ginc.GetInt(c, "status", 2)
	keyword := ginc.GetString(c, "keyword")
	guides, count, err := model.NewGuide().FindGuideAdminList(page, size, status, keyword)
	if err != nil {
		log.Get(c).Errorf("GuideList FindGuideAdminList error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	gameIds := make([]int32, 0, len(guides))
	for _, guide := range guides {
		gameIds = append(gameIds, guide.GameID)
	}
	//查找游戏相关数据
	gameMap, err := model.NewGames().FindGameMapByIds(gameIds)
	if err != nil {
		log.Get(c).Errorf("GuideList FindGameMapByIds error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	guideInfos := make([]*GuideInfo, 0, len(guides))
	for _, guide := range guides {
		gameName := ""
		if v, ok := gameMap[guide.GameID]; ok {
			gameName = v.GameName
		}
		guideInfos = append(guideInfos, &GuideInfo{
			Id:          guide.ID,
			GameId:      guide.GameID,
			Title:       guide.Title,
			Describe:    guide.Describe,
			URL:         guide.URL,
			Author:      guide.Author,
			AuthorImage: guide.AuthorImage,
			GameName:    gameName,
			Status:      guide.Status,
		})
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), guideInfos)
	ginc.Ok(c, res)
}

//GuideDelete 攻略删除
func GuideDelete(c *gin.Context) {
	guideId := ginc.GetInt32(c, "guide_id", 0)
	if guideId <= 0 {
		ginc.Fail(c, "please input guide_id", "400")
		return
	}
	datas := map[string]interface{}{
		"is_del": 1,
	}
	err := model.NewGuide().UpdateGuideInfoByGuideId(guideId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GuideStatus 攻略上下架
func GuideStatus(c *gin.Context) {
	guideId := ginc.GetInt32(c, "guide_id", 0)
	if guideId <= 0 {
		ginc.Fail(c, "please input guide_id", "400")
		return
	}
	status := ginc.GetInt(c, "status", 0)
	datas := map[string]interface{}{
		"status": status,
	}
	err := model.NewGuide().UpdateGuideInfoByGuideId(guideId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GuideCreate 新增攻略
func GuideCreate(c *gin.Context) {
	var v validate.GuideSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//新增攻略
	guide := &model.Guide{}
	guide.GameID = v.GameId
	guide.Title = v.Title
	guide.Describe = v.Describe
	guide.Content = v.Content
	guide.URL = v.Url
	guide.Author = v.Author
	guide.AuthorImage = v.AuthorImage
	guide.CreatedAt = time.Now()
	guide.UpdatedAt = time.Now()
	err := guide.Create()
	if err != nil {
		log.Get(c).Errorf("GuideCreate Create error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//GuideEdit 攻略编辑页
func GuideEdit(c *gin.Context) {
	guideId := ginc.GetInt(c, "id", 0)
	if guideId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	guide, err := model.NewGuide().FindInfoById(guideId, true)
	if err != nil {
		log.Get(c).Errorf("GuideEdit FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, guide)
}

//GuideUpdate 攻略更新
func GuideUpdate(c *gin.Context) {
	guideId := ginc.GetInt(c, "id", 0)
	if guideId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	var v validate.GuideSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//攻略更新
	guide, err := model.NewGuide().FindInfoById(guideId, true)
	if err != nil {
		log.Get(c).Errorf("GuideUpdate FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	guide.GameID = v.GameId
	guide.Title = v.Title
	guide.Describe = v.Describe
	guide.Content = v.Content
	guide.URL = v.Url
	guide.Author = v.Author
	guide.AuthorImage = v.AuthorImage
	err = guide.Update()
	if err != nil {
		log.Get(c).Errorf("GuideUpdate Update error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}
