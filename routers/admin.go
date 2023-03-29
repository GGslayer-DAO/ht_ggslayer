package routers

import (
	"ggslayer/controllers/admin"
	"ggslayer/middleware"
	"github.com/gin-gonic/gin"
)

//后台管理api
func adm(r *gin.Engine) {
	gadmin := r.Group("/api/adm")
	gadmin.POST("/admin/create", middleware.AdmAuth, admin.Create) //创建管理员
	gadmin.POST("/admin/login", admin.Login)                       //管理员登陆
	gadmin.POST("/file/upload", middleware.AdmAuth, admin.Upload)  //后台文件上传

	gadmin.GET("/game/list", middleware.AdmAuth, admin.GameList)                      //游戏列表
	gadmin.DELETE("/game/delete", middleware.AdmAuth, admin.GameDelete)               //游戏删除
	gadmin.GET("/game/status", middleware.AdmAuth, admin.GameStatus)                  //游戏上下架
	gadmin.POST("/game/term_investor", middleware.AdmAuth, admin.GameTermAndInvestor) //增加管理员和投资者
	gadmin.POST("/game/create", middleware.AdmAuth, admin.GameCreate)                 //游戏新增
	gadmin.GET("/game/edit/:id", middleware.AdmAuth, admin.GameEdit)                  //游戏编辑页
	gadmin.PUT("/game/update/:id", middleware.AdmAuth, admin.GameUpdate)              //游戏更新
	gadmin.GET("/game/add_trend", middleware.AdmAuth, admin.GameAddTrend)             //增加游戏趋势
	gadmin.GET("/game/get_all", middleware.AdmAuth, admin.GameGetAll)                 //查找所有游戏数据信息

	gadmin.GET("/guide/list", middleware.AdmAuth, admin.GuideList)         //攻略列表
	gadmin.DELETE("/guide/delete", middleware.AdmAuth, admin.GuideDelete)  //攻略删除
	gadmin.GET("/guide/status", middleware.AdmAuth, admin.GuideStatus)     //攻略上下架
	gadmin.POST("/guide/create", middleware.AdmAuth, admin.GuideCreate)    //新增攻略
	gadmin.GET("/guide/edit/:id", middleware.AdmAuth, admin.GuideEdit)     //攻略编辑页
	gadmin.PUT("/guide/update/:id", middleware.AdmAuth, admin.GuideUpdate) //攻略更新

	gadmin.GET("/event/list", middleware.AdmAuth, admin.EventList)         //活动列表
	gadmin.POST("/event/create", middleware.AdmAuth, admin.EventCreate)    //活动新增
	gadmin.DELETE("/event/delete", middleware.AdmAuth, admin.EventDelete)  //活动删除
	gadmin.GET("/event/status", middleware.AdmAuth, admin.EventStatus)     //活动上下架
	gadmin.GET("/event/edit/:id", middleware.AdmAuth, admin.EventEdit)     //活动编辑页
	gadmin.PUT("/event/update/:id", middleware.AdmAuth, admin.EventUpdate) //活动更新

	gadmin.GET("/contest/list", middleware.AdmAuth, admin.ContestList)         //比赛列表
	gadmin.POST("/contest/create", middleware.AdmAuth, admin.ContestCreate)    //比赛新增
	gadmin.DELETE("/contest/delete", middleware.AdmAuth, admin.ContestDelete)  //比赛删除
	gadmin.GET("/contest/status", middleware.AdmAuth, admin.ContestStatus)     //比赛状态变更
	gadmin.GET("/contest/edit/:id", middleware.AdmAuth, admin.ContestEdit)     //比赛编辑页
	gadmin.PUT("/contest/update/:id", middleware.AdmAuth, admin.ContestUpdate) //比赛更新
}
