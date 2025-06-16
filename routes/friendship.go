package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/middlewares/validators"
	"github.com/gin-gonic/gin"
)

func FriendshipRoute(router *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	friends := router.Group("/friends", handlers...)
	{
		friends.GET(
			"",
			controllers.GetFriends,
		)

		friends.POST(
			"/request",
			validators.SendFriendRequestValidator(),
			controllers.SendFriendRequest,
		)

		friends.POST(
			"/request/:id/respond",
			validators.PathIdValidator(),
			validators.RespondFriendRequestValidator(),
			controllers.RespondToFriendRequest,
		)

		friends.GET(
			"/requests/received",
			controllers.GetPendingFriendRequests,
		)

		friends.GET(
			"/requests/sent",
			controllers.GetSentFriendRequests,
		)

		friends.DELETE(
			"/:friendId",
			validators.PathIdValidator(),
			controllers.RemoveFriend,
		)

		friends.POST(
			"/block/:userId",
			validators.PathIdValidator(),
			controllers.BlockUser,
		)
	}
}
