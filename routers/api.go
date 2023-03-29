package routers

import (
	"ggslayer/controllers/contest"
	"ggslayer/controllers/event"
	"ggslayer/controllers/game"
	"ggslayer/controllers/game/meta"
	"ggslayer/controllers/guide"
	"ggslayer/controllers/home"
	"ggslayer/controllers/user"
	"ggslayer/middleware"
	"github.com/gin-gonic/gin"
)

func api(r *gin.Engine) {
	guser := r.Group("/api/gg_user")

	guser.POST("/post_email", user.PostEmail)     //发送邮件
	guser.POST("/register", user.Register)        //用户注册
	guser.POST("/code_login", user.CodeLogin)     //验证码登陆
	guser.POST("/pass_login", user.PassLogin)     //密码登陆
	guser.POST("/google_login", user.GoogleLogin) //谷歌登陆
	guser.POST("/reset_first", user.ResetFirst)   //重置输入邮箱确认
	guser.POST("/reset_pass", user.ResetPass)     //重置密码

	guser.GET("/logout", middleware.Auth, user.Logout) //用户退出功能

	guser.GET("/user_info", middleware.Auth, user.UserInfo)           //用户数据信息
	guser.POST("/user_update", middleware.Auth, user.UserUpdate)      //用户数据更新
	guser.POST("/to_reset_pass", middleware.Auth, user.ToResetPass)   //用户重置密码
	guser.POST("/change_email", middleware.Auth, user.ChangeEmail)    //用户更换邮箱
	guser.GET("/game_collect", middleware.Auth, user.UserGameCollect) //用户游戏收藏
	guser.GET("/game_vote", middleware.Auth, user.UserGameVote)       //用户游戏投票列表
	guser.GET("/user_rate", middleware.Auth, user.UserComment)        //用户评论列表
	guser.POST("/upload", middleware.Auth, user.Upload)
	guser.POST("/bind_token", middleware.Auth, user.BindAddressToken) //绑定钱包地址
	//guser.GET("/address_nft", middleware.Auth, user.FindBalanceNft)       //用户nft
	//guser.GET("/address_token", middleware.Auth, user.FindBalanceToken)   //用户地址token
	guser.GET("/address_access", middleware.Auth, user.FindBalanceAssets) //用户地址资产

	//游戏模块
	ggame := r.Group("/api/gg_game")
	//爬取游戏
	ggame.GET("/craw_meta_add", meta.CrawMetaGameAdd)           //爬取meta游戏
	ggame.GET("/change_chain_name", meta.ChangeChainName)       //修改简称
	ggame.GET("/contract_address_add", meta.ContractAddressAdd) //添加智能合约接口
	ggame.GET("/craw_meta_update", meta.CrawMetaGameUpdate)     //数据更新
	ggame.GET("/crow_dap", game.DapGameCraw)                    //爬取dap游戏

	ggame.GET("/home/game_tab", home.ShowHomeGameTab)            //首页游戏tag
	ggame.GET("/home/game_vote_list", home.ShowHomeGameVoteList) //展示首页游戏投票列表
	ggame.GET("/game_ranking", middleware.MustNotAuth, game.RankingGame)
	ggame.GET("/game_all_tag", game.AllTags)                                //获取所有游戏标签
	ggame.GET("/game_all_chain", game.AllChains)                            //获取所有游戏链
	ggame.GET("/game_collect", middleware.Auth, game.CollectGame)           //游戏收藏
	ggame.GET("/vote_is_can", middleware.Auth, game.VoteIsCan)              //判断游戏是否可投票和转发
	ggame.GET("/game_vote", middleware.Auth, game.VoteToGame)               //游戏投票
	ggame.GET("/game_forward", middleware.Auth, game.ForwardToGame)         //游戏转发
	ggame.GET("/game_detail", middleware.MustNotAuth, game.GameDetail)      //游戏详情数据
	ggame.GET("/game_kol", game.KolGameInfo)                                //获取game_kol数据信息
	ggame.GET("/game_token", game.TokenInfo)                                //获取game_token数据信息
	ggame.GET("/game_nft", game.NftInfo)                                    //获取游戏nft数据信息
	ggame.POST("/game_rate", middleware.Auth, game.GameRate)                //游戏评论
	ggame.GET("/game_rate_list", middleware.MustNotAuth, game.GameRateList) //游戏评分展示
	ggame.GET("/rate_thumb", middleware.Auth, game.RateThumb)               //评论点赞
	ggame.GET("/rate_statics", game.GameRateStatics)                        //游戏评分数据统计
	ggame.GET("/game_test", game.Cest)

	//攻略模块
	gguide := r.Group("/api/gg_guide")
	gguide.GET("/game_guide", middleware.MustNotAuth, guide.GameGuideList) //查找游戏攻略
	gguide.GET("/detail/:id", middleware.MustNotAuth, guide.GuideDetail)   //攻略详情
	gguide.GET("/guide_collect", middleware.Auth, guide.CollectGuide)      //攻略收藏

	//活动模块
	ggevent := r.Group("/api/gg_event")
	ggevent.GET("/show_event_list", event.ShowEventList) //活动页面展示

	ggcontest := r.Group("/api/gg_contest")
	ggcontest.GET("/contest_info", middleware.MustNotAuth, contest.IndexInfo) //当前比赛列表
	ggcontest.GET("/contest_title", contest.IndexTitle)                       //获取往期比赛标题
}
