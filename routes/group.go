package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/middlewares/validators"
	"github.com/gin-gonic/gin"
)

func GroupRoute(router *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	groups := router.Group("/groups", handlers...)
	{
		groups.POST(
			"",
			validators.CreateGroupValidator(),
			controllers.CreateGroup,
		)

		groups.GET(
			"",
			controllers.GetUserGroups,
		)

		groups.GET(
			"/:id",
			validators.PathIdValidator(),
			controllers.GetGroupById,
		)

		groups.PUT(
			"/:id",
			validators.PathIdValidator(),
			validators.UpdateGroupValidator(),
			controllers.UpdateGroup,
		)

		groups.DELETE(
			"/:id",
			validators.PathIdValidator(),
			controllers.DeleteGroup,
		)

		// Member management
		groups.POST(
			"/:id/members",
			validators.PathIdValidator(),
			validators.AddMemberToGroupValidator(),
			controllers.AddMemberToGroup,
		)

		groups.GET(
			"/:id/members",
			validators.PathIdValidator(),
			controllers.GetGroupMembers,
		)

		groups.DELETE(
			"/:id/members/:memberId",
			validators.PathIdValidator(),
			controllers.RemoveMemberFromGroup,
		)

		// Note: Group expenses now available via /v1/groups/:id/transactions/expenses
		// Note: Balances and settlement routes moved to transaction routes
		// for the new unified transaction-based architecture
	}
}
