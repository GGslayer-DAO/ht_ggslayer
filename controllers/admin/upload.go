package admin

import (
	"fmt"
	"ggslayer/utils"
	"ggslayer/utils/config"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"path"
	"time"
)

//Upload 上传图片
func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "upload fail", "400")
		return
	}
	fileHandle, err := file.Open() //打开上传文件
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "upload fail", "400")
		return
	}
	defer fileHandle.Close()
	fileByte, err := ioutil.ReadAll(fileHandle) //获取上传文件字节流
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, "upload fail", "400")
		return
	}
	fileExt := path.Ext(file.Filename) //获取文件后缀名
	fileNameStr := utils.Md5V(fmt.Sprintf("%s%s", file.Filename, time.Now().String()))
	fileName := fmt.Sprintf("%s%s", fileNameStr, fileExt)
	client := utils.NewClient()
	err = client.UploadImage(fileName, fileByte)
	if err != nil {
		ginc.Fail(c, "upload fail", "400")
		return
	}
	ginc.Ok(c, config.GetString("oss.Domain")+fileName)
}
