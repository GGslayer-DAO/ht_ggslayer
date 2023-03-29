package admin

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"strings"
	"time"
)

//EventList 活动列表
func EventList(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	status := ginc.GetInt(c, "status", 2)
	keyword := ginc.GetString(c, "keyword")
	ty := ginc.GetString(c, "type")
	events, count, err := model.NewEvent().FindEventList(page, size, status, keyword, ty, true)
	if err != nil {
		log.Get(c).Errorf("EventList FindEventList error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), events)
	ginc.Ok(c, res)
}

//EventDelete 活动删除
func EventDelete(c *gin.Context) {
	eventId := ginc.GetInt32(c, "event_id", 0)
	if eventId <= 0 {
		ginc.Fail(c, "please input event_id", "400")
		return
	}
	datas := map[string]interface{}{
		"is_del": 1,
	}
	err := model.NewEvent().UpdateEventInfoByEventId(eventId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//EventStatus 活动上下架
func EventStatus(c *gin.Context) {
	eventId := ginc.GetInt32(c, "event_id", 0)
	if eventId <= 0 {
		ginc.Fail(c, "please input event_id", "400")
		return
	}
	status := ginc.GetInt(c, "status", 0)
	datas := map[string]interface{}{
		"status": status,
	}
	err := model.NewEvent().UpdateEventInfoByEventId(eventId, datas)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, "success")
}

//EventCreate 活动新增
func EventCreate(c *gin.Context) {
	var v validate.EventSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//交易各类活动参数
	err := checkEvent(v)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx := model.GetDB().Begin()
	//新增活动
	event := &model.Event{}
	event.Name = v.Name
	event.Describe = v.Describe
	event.URL = v.Url
	event.Image = v.Image
	event.Tags = strings.Join(v.Tags, ",")
	event.EndTime, _ = time.ParseInLocation(utils.TimeLayoutStr, v.EndTime, time.Local)
	event.Type = v.Type
	err = tx.Save(event).Error
	if err != nil {
		log.Get(c).Errorf("EventCreate Save error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	eventId := event.ID
	err = eventSave(v, eventId, tx)
	if err != nil {
		log.Get(c).Errorf("EventCreate eventSave error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx.Commit()
	ginc.Ok(c, "success")
}

func checkEvent(v validate.EventSaveValidate) (err error) {
	if v.Type == "IGO" || v.Type == "INO" {
		if v.TotalAward == 0 {
			err = fmt.Errorf("the total_award cannot be empty")
			return
		}
		if v.Price == 0 {
			err = fmt.Errorf("the price cannot be empty")
			return
		}
	}
	if v.Type == "INO" {
		if v.TotalCount == 0 {
			err = fmt.Errorf("the total_count cannot be empty")
			return
		}
	}
	if v.Type == "Airdrop" {
		if v.AirdropReward == "" {
			err = fmt.Errorf("the airdrop_reward cannot be empty")
			return
		}
		if v.Quota == 0 {
			err = fmt.Errorf("the quota cannot be empty")
			return
		}
	}
	return
}

//EventEdit 活动编辑页
func EventEdit(c *gin.Context) {
	eventId := ginc.GetInt32(c, "id", 0)
	if eventId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	event, err := model.NewEvent().FindInfoById(eventId)
	if err != nil {
		log.Get(c).Errorf("EventEdit FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ty := event.Type
	if ty == "IGO" {
		igo, er := model.NewEventIgo().FindInfoByEventId(eventId)
		if er != nil {
			log.Get(c).Errorf("EventEdit eventigo FindInfoByEventId error(%s)", err.Error())
			ginc.Fail(c, err.Error(), "400")
			return
		}
		event.EventDetail = igo
	} else if ty == "INO" {
		ino, er := model.NewEventIno().FindInfoByEventId(eventId)
		if er != nil {
			log.Get(c).Errorf("EventEdit eventino FindInfoByEventId error(%s)", err.Error())
			ginc.Fail(c, err.Error(), "400")
			return
		}
		event.EventDetail = ino
	} else {
		airdrop, er := model.NewEventAirdrop().FindInfoByEventId(eventId)
		if er != nil {
			log.Get(c).Errorf("EventEdit eventairdrop FindInfoByEventId error(%s)", err.Error())
			ginc.Fail(c, err.Error(), "400")
			return
		}
		event.EventDetail = airdrop
	}
	ginc.Ok(c, event)
}

//EventUpdate 活动更新
func EventUpdate(c *gin.Context) {
	eventId := ginc.GetInt32(c, "id", 0)
	if eventId <= 0 {
		ginc.Fail(c, "page error", "400")
		return
	}
	var v validate.EventSaveValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	//活动更新
	event, err := model.NewEvent().FindInfoById(eventId)
	if err != nil {
		log.Get(c).Errorf("EventUpdate FindInfoById error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx := model.GetDB().Begin()
	event.Name = v.Name
	event.Describe = v.Describe
	event.URL = v.Url
	event.Image = v.Image
	event.Tags = strings.Join(v.Tags, ",")
	event.EndTime, _ = time.ParseInLocation(utils.TimeLayoutStr, v.EndTime, time.Local)
	event.Type = v.Type
	err = tx.Save(event).Error
	if err != nil {
		log.Get(c).Errorf("EventUpdate Save error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	eventId = event.ID
	err = eventSave(v, eventId, tx)
	if err != nil {
		log.Get(c).Errorf("EventUpdate eventSave error(%s)", err.Error())
		tx.Rollback()
		ginc.Fail(c, err.Error(), "400")
		return
	}
	tx.Commit()
	ginc.Ok(c, "success")
}

func eventSave(v validate.EventSaveValidate, eventId int32, tx *gorm.DB) (err error) {
	if v.Type == "IGO" {
		igo, _ := model.NewEventIgo().FindInfoByEventId(eventId)
		igo.EventID = eventId
		igo.TotalAward = v.TotalAward
		igo.Price = v.Price
		err = tx.Save(igo).Error
	} else if v.Type == "INO" {
		ino, _ := model.NewEventIno().FindInfoByEventId(eventId)
		ino.EventID = eventId
		ino.TotalAward = v.TotalAward
		ino.Price = v.Price
		ino.TotalCount = v.TotalCount
		err = tx.Save(ino).Error
	} else {
		airdrop, _ := model.NewEventAirdrop().FindInfoByEventId(eventId)
		airdrop.EventID = eventId
		airdrop.AirdropReward = v.AirdropReward
		airdrop.Quota = v.Quota
		err = tx.Save(airdrop).Error
	}
	return
}
