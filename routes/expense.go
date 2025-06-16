package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/middlewares/validators"
	"github.com/gin-gonic/gin"
)

func ExpenseRoute(router *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	expenses := router.Group("/expenses", handlers...)
	{
		expenses.POST(
			"",
			validators.CreateExpenseValidator(),
			controllers.CreateExpense,
		)

		expenses.GET(
			"",
			controllers.GetUserExpenses,
		)

		expenses.GET(
			"/:id",
			validators.PathIdValidator(),
			controllers.GetExpenseById,
		)

		expenses.PUT(
			"/:id",
			validators.PathIdValidator(),
			validators.UpdateExpenseValidator(),
			controllers.UpdateExpense,
		)

		expenses.DELETE(
			"/:id",
			validators.PathIdValidator(),
			controllers.DeleteExpense,
		)
	}
}
