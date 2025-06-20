package controllers

import (
	"net/http"
	"strconv"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var transactionService = &services.TransactionService{}

// CreateExpense godoc
// @Summary      Create Expense (New Transaction Model)
// @Description  creates a new expense using the unified transaction model
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        expense  body      models.CreateExpenseTransactionRequest  true  "Expense Request"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /transactions/expense [post]
// @Security     ApiKeyAuth
func CreateExpenseTransaction(c *gin.Context) {
	var requestBody models.CreateExpenseTransactionRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	transaction, err := transactionService.CreateExpenseTransaction(userId.(primitive.ObjectID), requestBody)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Success = true
	response.Data = gin.H{"transaction": transaction}
	response.Message = "Expense created successfully"
	response.SendResponse(c)
}

// CreateSettlement godoc
// @Summary      Create Settlement (New Transaction Model)
// @Description  creates a settlement using the unified transaction model
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        settlement  body      models.CreateSettlementTransactionRequest  true  "Settlement Request"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /transactions/settlement [post]
// @Security     ApiKeyAuth
func CreateSettlementTransaction(c *gin.Context) {
	var requestBody models.CreateSettlementTransactionRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	groupId, err := primitive.ObjectIDFromHex(requestBody.GroupID)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	payerId, err := primitive.ObjectIDFromHex(requestBody.PayerID)
	if err != nil {
		response.Message = "invalid payer id"
		response.SendResponse(c)
		return
	}

	payeeId, err := primitive.ObjectIDFromHex(requestBody.PayeeID)
	if err != nil {
		response.Message = "invalid payee id"
		response.SendResponse(c)
		return
	}

	transaction, err := transactionService.CreateSettlementTransaction(groupId, payerId, payeeId, requestBody.Amount, requestBody.Currency, requestBody.Notes, requestBody.IsCompleted, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Success = true
	response.Data = gin.H{"transaction": transaction}
	response.Message = "Settlement created successfully"
	response.SendResponse(c)
}

// GetGroupBalances godoc
// @Summary      Get Group Balances
// @Description  gets real-time balance summary for a group using maintained balances
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/balances [get]
// @Security     ApiKeyAuth
func GetGroupBalancesV2(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	balances, err := transactionService.GetGroupBalances(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"balances": balances}
	response.SendResponse(c)
}

// SimplifyDebts godoc
// @Summary      Simplify Group Debts
// @Description  calculates simplified settlement suggestions using maintained balances
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/simplify [get]
// @Security     ApiKeyAuth
func SimplifyDebtsV2(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	settlements, err := transactionService.SimplifyDebtsFromBalances(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"suggested_settlements": settlements}
	response.SendResponse(c)
}

// MarkTransactionComplete godoc
// @Summary      Mark Transaction Complete
// @Description  marks a settlement transaction as completed
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Param        req  body      models.CompleteTransactionRequest false "Complete Transaction Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /transactions/{id}/complete [post]
// @Security     ApiKeyAuth
func MarkTransactionComplete(c *gin.Context) {
	var requestBody models.CompleteTransactionRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	transactionId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid transaction id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = transactionService.MarkTransactionComplete(transactionId, userId.(primitive.ObjectID), requestBody.Notes, requestBody.SettlementMethod, requestBody.ProofOfPayment)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Transaction marked as complete"
	response.SendResponse(c)
}

// GetGroupTransactions godoc
// @Summary      Get Group Transactions
// @Description  gets all transactions for a specific group
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Param        type     query     string  false  "Filter by transaction type (expense, settlement)"
// @Param        page     query     int     false  "Page number (default: 0)"
// @Param        limit    query     int     false  "Items per page (default: 20)"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/transactions [get]
// @Security     ApiKeyAuth
func GetGroupTransactions(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	// Parse query parameters
	transactionType := c.Query("type")

	page := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	transactions, err := transactionService.GetGroupTransactions(groupId, userId.(primitive.ObjectID), transactionType, page, limit)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	// Check if there are more results
	hasMore := len(transactions) > limit
	if hasMore {
		transactions = transactions[:limit] // Remove the extra item
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{
		"transactions": transactions,
		"page":         page,
		"limit":        limit,
		"has_more":     hasMore,
	}
	response.SendResponse(c)
}

// GetTransactionById godoc
// @Summary      Get Transaction by ID
// @Description  gets a single transaction by ID
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /transactions/{id} [get]
// @Security     ApiKeyAuth
func GetTransactionById(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	transactionId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid transaction id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	transaction, err := transactionService.GetTransactionById(transactionId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"transaction": transaction}
	response.SendResponse(c)
}

// UpdateTransaction godoc
// @Summary      Update Transaction
// @Description  updates an expense transaction
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Param        req  body      models.UpdateTransactionRequest  true  "Update Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /transactions/{id} [put]
// @Security     ApiKeyAuth
func UpdateTransaction(c *gin.Context) {
	var requestBody models.UpdateTransactionRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	transactionId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid transaction id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = transactionService.UpdateTransaction(transactionId, userId.(primitive.ObjectID), requestBody)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Transaction updated successfully"
	response.SendResponse(c)
}

// DeleteTransaction godoc
// @Summary      Delete Transaction
// @Description  deletes a transaction
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /transactions/{id} [delete]
// @Security     ApiKeyAuth
func DeleteTransaction(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	transactionId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid transaction id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = transactionService.DeleteTransaction(transactionId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Transaction deleted successfully"
	response.SendResponse(c)
}

// GetGroupExpenseTransactions godoc
// @Summary      Get Group Expense Transactions
// @Description  gets all expense transactions for a group
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Param        page     query     int     false  "Page number (default: 0)"
// @Param        limit    query     int     false  "Items per page (default: 20)"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/transactions/expenses [get]
// @Security     ApiKeyAuth
func GetGroupExpenseTransactions(c *gin.Context) {
	getGroupTransactionsByType(c, "expense")
}

// GetGroupSettlementTransactions godoc
// @Summary      Get Group Settlement Transactions
// @Description  gets all settlement transactions for a group
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Param        page     query     int     false  "Page number (default: 0)"
// @Param        limit    query     int     false  "Items per page (default: 20)"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/transactions/settlements [get]
// @Security     ApiKeyAuth
func GetGroupSettlementTransactions(c *gin.Context) {
	getGroupTransactionsByType(c, "settlement")
}

// getGroupTransactionsByType is a helper function for filtering transactions by type
func getGroupTransactionsByType(c *gin.Context, transactionType string) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	// Parse query parameters
	page := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	transactions, err := transactionService.GetGroupTransactions(groupId, userId.(primitive.ObjectID), transactionType, page, limit)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	// Check if there are more results
	hasMore := len(transactions) > limit
	if hasMore {
		transactions = transactions[:limit]
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{
		"transactions": transactions,
		"type":         transactionType,
		"page":         page,
		"limit":        limit,
		"has_more":     hasMore,
	}
	response.SendResponse(c)
}

// GetGroupBalanceHistory godoc
// @Summary      Get Group Balance History
// @Description  gets balance change history for a group
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Param        days     query     int     false  "Number of days (default: 30)"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/balance-history [get]
// @Security     ApiKeyAuth
func GetGroupBalanceHistory(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	history, err := transactionService.GetGroupBalanceHistory(groupId, userId.(primitive.ObjectID), days)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"balance_history": history}
	response.SendResponse(c)
}

// GetGroupAnalytics godoc
// @Summary      Get Group Analytics
// @Description  gets analytics data for a group
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/analytics [get]
// @Security     ApiKeyAuth
func GetGroupAnalytics(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	analytics, err := transactionService.GetGroupAnalytics(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"analytics": analytics}
	response.SendResponse(c)
}

// CreateBulkSettlements godoc
// @Summary      Create Bulk Settlements
// @Description  creates multiple settlements from suggested settlements
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Param        req      body      models.BulkSettlementsTransactionRequest  true  "Bulk Settlements Request"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/bulk-settlements [post]
// @Security     ApiKeyAuth
func CreateBulkSettlements(c *gin.Context) {
	var requestBody models.BulkSettlementsTransactionRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	settlements, err := transactionService.CreateBulkSettlements(groupId, userId.(primitive.ObjectID), requestBody.Settlements)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Success = true
	response.Data = gin.H{"settlements": settlements}
	response.Message = "Bulk settlements created successfully"
	response.SendResponse(c)
}

// RecalculateGroupBalances godoc
// @Summary      Recalculate Group Balances
// @Description  recalculates all balances for a group (admin operation)
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/recalculate-balances [post]
// @Security     ApiKeyAuth
func RecalculateGroupBalances(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	groupIdHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(groupIdHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = transactionService.RecalculateGroupBalances(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Group balances recalculated successfully"
	response.SendResponse(c)
}

// GetUserTransactions godoc
// @Summary      Get User Transactions
// @Description  gets all transactions for the authenticated user across all groups
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number (default: 0)"
// @Param        limit    query     int     false  "Items per page (default: 20)"
// @Param        type     query     string  false  "Filter by transaction type"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /users/me/transactions [get]
// @Security     ApiKeyAuth
func GetUserTransactions(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	// Parse query parameters
	transactionType := c.Query("type")

	page := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	transactions, err := transactionService.GetUserTransactions(userId.(primitive.ObjectID), transactionType, page, limit)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	// Check if there are more results
	hasMore := len(transactions) > limit
	if hasMore {
		transactions = transactions[:limit]
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{
		"transactions": transactions,
		"page":         page,
		"limit":        limit,
		"has_more":     hasMore,
	}
	response.SendResponse(c)
}

// GetUserBalances godoc
// @Summary      Get User Balances
// @Description  gets all group balances for the authenticated user
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /users/me/balances [get]
// @Security     ApiKeyAuth
func GetUserBalances(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	balances, err := transactionService.GetUserBalances(userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"balances": balances}
	response.SendResponse(c)
}

// GetUserAnalytics godoc
// @Summary      Get User Analytics
// @Description  gets analytics data for the authenticated user
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /users/me/analytics [get]
// @Security     ApiKeyAuth
func GetUserAnalytics(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	analytics, err := transactionService.GetUserAnalytics(userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"analytics": analytics}
	response.SendResponse(c)
}
