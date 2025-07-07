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
			"/profile",
			controllers.UpdateProfile,
		)

		// Push notification subscription endpoints
		user.POST(
			"/push-subscription",
			controllers.RegisterPushSubscription,
		)
		user.PUT(
			"/push-subscription/:id",
			controllers.UpdatePushSubscription,
		)
		user.GET(
			"/push-subscriptions",
			controllers.GetPushSubscriptions,
		)
		user.DELETE(
			"/push-subscription/:id",
			controllers.DeregisterPushSubscription,
		)
		user.DELETE(
			"/push-subscriptions",
			controllers.DeregisterAllPushSubscriptions,
		)
		user.POST(
			"/push-subscription/:id/test",
			controllers.TestPushNotification,
		)
	}
}
