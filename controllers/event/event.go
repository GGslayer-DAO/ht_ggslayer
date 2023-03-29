package event

import (
	"ggslayer/model"
	"ggslayer/utils"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"math"
)

//ShowEventList 展示活动列表
func ShowEventList(c *gin.Context) {
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	ty := ginc.GetString(c, "type")
	events, count, err := model.NewEvent().FindEventList(page, size, 2, "", ty, false)
	if err != nil {
		log.Get(c).Errorf("ShowEventList FindEventList error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//收集各种活动
	igoArr := make([]int32, 0, len(events))
	inoArr := make([]int32, 0, len(events))
	airdropArr := make([]int32, 0, len(events))
	for _, event := range events {
		if event.Type == "IGO" {
			igoArr = append(igoArr, event.ID)
		} else if event.Type == "INO" {
			inoArr = append(inoArr, event.ID)
		} else {
			airdropArr = append(airdropArr, event.ID)
		}
	}
	igoMap := make(map[int32]*model.EventIgo)
	inoMap := make(map[int32]*model.EventIno)
	airdropMap := make(map[int32]*model.EventAirdrop)
	//关联查询
	eg := errgroup.Group{}
	eg.Go(func() error {
		if len(igoArr) == 0 {
			return nil
		}
		igoMap, err = model.NewEventIgo().FindInfoByEventIds(igoArr)
		return err
	})
	eg.Go(func() error {
		if len(inoArr) == 0 {
			return nil
		}
		inoMap, err = model.NewEventIno().FindInfoByEventIds(inoArr)
		return err
	})
	eg.Go(func() error {
		if len(airdropArr) == 0 {
			return nil
		}
		airdropMap, err = model.NewEventAirdrop().FindInfoByEventIds(airdropArr)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(c).Errorf("ShowEventList error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	for _, event := range events {
		if v, ok := igoMap[event.ID]; ok {
			event.EventDetail = v
		}
		if v, ok := inoMap[event.ID]; ok {
			event.EventDetail = v
		}
		if v, ok := airdropMap[event.ID]; ok {
			event.EventDetail = v
		}
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), events)
	ginc.Ok(c, res)
}
