package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/middlewares/validators"
	"github.com/gin-gonic/gin"
)

func SettlementRoute(router *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	settlements := router.Group("/settlements", handlers...)
	{
		settlements.GET(
			"",
			controllers.GetUserSettlements,
		)

		settlements.POST(
			"/:id/complete",
			validators.PathIdValidator(),
			validators.SettleDebtValidator(),
			controllers.MarkSettlementComplete,
		)
	}
}
