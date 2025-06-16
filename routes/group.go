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

		// Expenses in groups
		groups.GET(
			"/:id/expenses",
			validators.PathIdValidator(),
			controllers.GetGroupExpenses,
		)

		// Balances and settlements
		groups.GET(
			"/:id/balances",
			validators.PathIdValidator(),
			controllers.GetGroupBalances,
		)

		groups.GET(
			"/:id/simplify",
			validators.PathIdValidator(),
			controllers.SimplifyDebts,
		)

		groups.GET(
			"/:id/settlements",
			validators.PathIdValidator(),
			controllers.GetGroupSettlements,
		)
	}
}
