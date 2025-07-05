package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	user := router.Group("/user", handlers...)
	{
		user.GET(
			"/me",
			controllers.GetMe,
		)
		user.PUT(
			"/fcm-token",
			controllers.UpdateFCMToken,
		)
		user.PUT(
			"/profile",
			controllers.UpdateProfile,
		)
	}
}
