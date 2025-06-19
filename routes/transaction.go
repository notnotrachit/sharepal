package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/middlewares"
	"github.com/gin-gonic/gin"
)

// TransactionRoutes defines routes for the new transaction-based API
func TransactionRoutes(router *gin.RouterGroup) {
	// Transaction routes (new unified API)
	transactionGroup := router.Group("/transactions")
	transactionGroup.Use(middlewares.JWTMiddleware())
	{
		// Create transactions
		transactionGroup.POST("/expense", controllers.CreateExpenseTransaction)
		transactionGroup.POST("/settlement", controllers.CreateSettlementTransaction)

		// Complete transactions
		transactionGroup.POST("/:id/complete", controllers.MarkTransactionComplete)

		// Get single transaction
		transactionGroup.GET("/:id", controllers.GetTransactionById)

		// Update transaction (for editing expenses)
		transactionGroup.PUT("/:id", controllers.UpdateTransaction)

		// Delete transaction
		transactionGroup.DELETE("/:id", controllers.DeleteTransaction)
	}

	// Enhanced group routes with new balance/transaction endpoints
	groupGroup := router.Group("/groups")
	groupGroup.Use(middlewares.JWTMiddleware())
	{
		// Balance endpoints (replacing old implementation)
		groupGroup.GET("/:id/balances", controllers.GetGroupBalancesV2)
		groupGroup.GET("/:id/simplify", controllers.SimplifyDebtsV2)

		// Transaction endpoints
		groupGroup.GET("/:id/transactions", controllers.GetGroupTransactions)
		groupGroup.GET("/:id/transactions/expenses", controllers.GetGroupExpenseTransactions)
		groupGroup.GET("/:id/transactions/settlements", controllers.GetGroupSettlementTransactions)

		// Balance history and analytics
		groupGroup.GET("/:id/balance-history", controllers.GetGroupBalanceHistory)
		groupGroup.GET("/:id/analytics", controllers.GetGroupAnalytics)

		// Bulk operations
		groupGroup.POST("/:id/bulk-settlements", controllers.CreateBulkSettlements)
		groupGroup.POST("/:id/recalculate-balances", controllers.RecalculateGroupBalances)
	}

	// User transaction routes
	userGroup := router.Group("/users")
	userGroup.Use(middlewares.JWTMiddleware())
	{
		// User's transactions across all groups
		userGroup.GET("/me/transactions", controllers.GetUserTransactions)
		userGroup.GET("/me/balances", controllers.GetUserBalances)
		userGroup.GET("/me/analytics", controllers.GetUserAnalytics)
	}
}
