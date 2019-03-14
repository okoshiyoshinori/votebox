package router

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/okoshiyoshinori/votebox/config"
	"github.com/okoshiyoshinori/votebox/votetoken"
)

func GetRouter() *gin.Engine {

	r := gin.Default()

	//r.Static("/static/item", config.APConfig.Item)

	store := cookie.NewStore([]byte("voteSecret"))
	r.Use(sessions.Sessions("votesession", store))

	r.Use(CORSMiddleWare())

	privateApi := r.Group("/api/private/", PrivateMiddleWare())
	PrivateApiRouter(privateApi)

	publicApi := r.Group("/api/public/")
	PublicApiRouter(publicApi)

	ws := r.Group("/ws")
	WebsockRouter(ws)

	return r
}

func CORSMiddleWare() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.APConfig.Origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func PrivateMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		ok := votetoken.CheckToken(c)
		if !ok {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}
