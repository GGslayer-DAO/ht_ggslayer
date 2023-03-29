package user

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/service"
	"ggslayer/utils"
	"ggslayer/utils/config"
	"ggslayer/utils/ginc"
	"ggslayer/utils/log"
	"ggslayer/validate"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"math"
	"path"
	"time"
)

//UserInfo用户基础信息展示
func UserInfo(c *gin.Context) {
	userId := c.GetInt("user_id")
	user, err := model.NewUser().FindById(userId)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//查找钱包
	userToken, err := model.NewUserToken().FindTokenByUserId(userId)
	if err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, err.Error(), "400")
		return
	}
	user.UserToken = userToken
	ginc.Ok(c, user)
}

//Upload 上传图片
func Upload(c *gin.Context) {
	file, err := c.FormFile("img_file")
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
	if !utils.InArray(fileExt, []string{".png", ".jpg", ".gif", ".jpeg"}) {
		ginc.Fail(c, "upload fail, the file is must one of (.png, .jpg, .gif, .jpeg)", "400")
		return
	}
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

//UserUpdate 用户信息编辑
func UserUpdate(c *gin.Context) {
	var v validate.UserUpdateValidate
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Get(c).Error(err)
		ginc.Fail(c, v.GetError(err), "400")
		return
	}
	userId := c.GetInt("user_id")
	user, err := model.NewUser().FindById(userId)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	flag := 0
	if user.HeadPic == "" {
		flag = 1
	}
	user.Name = v.Name
	user.HeadPic = v.HeadPic
	user.Describe = v.Describe
	user.ID = int32(userId)
	err = user.Update()
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	//更换头像加20经验值
	if flag == 1 {
		service.AddExp(userId, 3, 20)
	}
	ginc.Ok(c, "success")
}

type GameInfo struct {
	Id       int32    `json:"id"`
	Name     string   `json:"name"`
	GameName string   `json:"game_name"`
	Logo     string   `json:"logo"`
	Tags     []string `json:"tags"`
	Chains   []string `json:"chains"`
	Motivate int      `json:"motivate"`
}

//UserGameCollect 我的游戏收藏
func UserGameCollect(c *gin.Context) {
	userId := c.GetInt("user_id")
	gameIds, err := model.NewUserGameCollect().FindUserCollectGameByUserId(userId)
	if err != nil {
		log.Get(c).Errorf("UserGameCollect FindUserCollectGameByUserId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	games, chainMap, tagMap, err := commonGameInfo(c, gameIds)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	gameInfos := make([]*GameInfo, 0, len(games))
	for _, game := range games {
		gameInfos = append(gameInfos, &GameInfo{
			Id:       game.ID,
			Name:     game.Name,
			GameName: game.GameName,
			Logo:     game.Logo,
			Tags:     tagMap[game.ID],
			Chains:   chainMap[game.ID],
		})
	}
	ginc.Ok(c, gameInfos)
}

//UserGameVote 用户游戏投票列表
func UserGameVote(c *gin.Context) {
	userId := c.GetInt("user_id")
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	gameIds, count, err := model.NewMyGameVote().FindVotePickByUserId(userId, page, size)
	if err != nil {
		log.Get(c).Errorf("UserGameVote FindVotePickByUserId error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
		return
	}
	if count == 0 {
		res := utils.PageReturn(page, size, 0, 0, []interface{}{})
		ginc.Ok(c, res)
		return
	}
	games, chainMap, tagMap, err := commonGameInfo(c, gameIds)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	gameInfo := make([]*GameInfo, 0, len(games))
	for _, game := range games {
		gameInfo = append(gameInfo, &GameInfo{
			Id:       game.ID,
			Name:     game.Name,
			GameName: game.GameName,
			Logo:     game.Logo,
			Tags:     tagMap[game.ID],
			Chains:   chainMap[game.ID],
			Motivate: int(game.Motivate),
		})
	}
	lastPage := math.Ceil(float64(count) / float64(size))
	res := utils.PageReturn(page, size, int(count), int(lastPage), gameInfo)
	ginc.Ok(c, res)
}

func commonGameInfo(c *gin.Context, gameIds []int32) (games []*model.Game, chainMap map[int32][]string,
	tagMap map[int32][]string, err error) {
	eg := errgroup.Group{}
	//查找游戏相关
	eg.Go(func() error {
		games, err = model.NewGames().FindGameInfoByGameIds(gameIds)
		return err
	})
	//查找链相关
	eg.Go(func() error {
		chainMap, err = model.NewGameBlockChain().FindChainByGameId(gameIds)
		return err
	})
	//查找标签相关
	eg.Go(func() error {
		tagMap, err = model.NewGameTag().FindTagsByGameIds(gameIds)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Get(c).Errorf("commonGameInfo findinfo error(%s)", err.Error())
		ginc.Fail(c, err.Error(), "400")
	}
	return
}

type UserRateInfo struct {
	ID          int32  `json:"id"`
	GameID      int32  `json:"game_id"`
	UserID      int32  `json:"user_id"`
	Rate        int    `json:"rate"`
	ExperReason string `json:"exper_reason"`
	UserName    string `json:"user_name"`
	UserPic     string `json:"user_pic"`
	CreatedAt   string `json:"created_at"`
	ThumbNumber int    `json:"thumb_number"`
}

//UserComment 用户评论
func UserComment(c *gin.Context) {
	userId := c.GetInt("user_id")
	page := ginc.GetInt(c, "page", 1)
	size := ginc.GetInt(c, "size", 10)
	s := service.NewGameRateService(c, userId, 0)
	res, err := s.FindMyRateInfo(page, size)
	if err != nil {
		ginc.Fail(c, err.Error(), "400")
		return
	}
	ginc.Ok(c, res)
}
