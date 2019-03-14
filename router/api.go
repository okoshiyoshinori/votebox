package router

import (
	"github.com/gin-gonic/gin"
	"github.com/okoshiyoshinori/votebox/controller"
)

func PrivateApiRouter(api *gin.RouterGroup) {
	api.GET("/user", controller.GetUserByIdController)
	api.POST("/box", controller.PostBoxController)
	api.PUT("/user", controller.PutUserController)
	api.DELETE("/user", controller.DeleteUserController)
	api.DELETE("/box/:bid", controller.DeleteBoxController)
}

func PublicApiRouter(api *gin.RouterGroup) {
	api.GET("/box", controller.GetBoxController)
	api.GET("/box/:bid", controller.GetBoxByIdController)
	api.POST("/user", controller.PostUserController)
	api.PUT("/vote", controller.PutVoteController)
	api.GET("/count", controller.GetBoxCount)
	api.GET("/oauth", controller.TwitterOauth)
	api.POST("/twittercallback", controller.CallbackController)
}

func WebsockRouter(api *gin.RouterGroup) {
	api.GET("/box/:bid", controller.WsServeController)
}
