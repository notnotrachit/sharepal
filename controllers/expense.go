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

// CreateExpense godoc
// @Summary      Create Expense
// @Description  creates a new expense in a group
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        req  body      models.CreateExpenseRequest true "Expense Request"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /expenses [post]
// @Security     ApiKeyAuth
func CreateExpense(c *gin.Context) {
	var requestBody models.CreateExpenseRequest
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

	expense, err := services.CreateExpense(userId.(primitive.ObjectID), requestBody)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Success = true
	response.Data = gin.H{"expense": expense}
	response.SendResponse(c)
}

// GetGroupExpenses godoc
// @Summary      Get Group Expenses
// @Description  gets expenses for a specific group with pagination
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        groupId  path     string  true  "Group ID"
// @Param        page     query    string  false  "Switch page by 'page'"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/expenses [get]
// @Security     ApiKeyAuth
func GetGroupExpenses(c *gin.Context) {
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

	pageQuery := c.DefaultQuery("page", "0")
	page, _ := strconv.Atoi(pageQuery)
	limit := 10

	expenses, err := services.GetGroupExpenses(groupId, userId.(primitive.ObjectID), page, limit)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	hasPrev := page > 0
	hasNext := len(expenses) > limit
	if hasNext {
		expenses = expenses[:limit]
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"expenses": expenses, "prev": hasPrev, "next": hasNext}
	response.SendResponse(c)
}

// GetUserExpenses godoc
// @Summary      Get User Expenses
// @Description  gets all expenses for the authenticated user with pagination
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        page  query    string  false  "Switch page by 'page'"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /expenses [get]
// @Security     ApiKeyAuth
func GetUserExpenses(c *gin.Context) {
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

	pageQuery := c.DefaultQuery("page", "0")
	page, _ := strconv.Atoi(pageQuery)
	limit := 10

	expenses, err := services.GetUserExpenses(userId.(primitive.ObjectID), page, limit)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	hasPrev := page > 0
	hasNext := len(expenses) > limit
	if hasNext {
		expenses = expenses[:limit]
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"expenses": expenses, "prev": hasPrev, "next": hasNext}
	response.SendResponse(c)
}

// GetExpenseById godoc
// @Summary      Get Expense
// @Description  get expense by id
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Expense ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /expenses/{id} [get]
// @Security     ApiKeyAuth
func GetExpenseById(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	expenseId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid expense id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	expense, err := services.GetExpenseById(expenseId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"expense": expense}
	response.SendResponse(c)
}

// UpdateExpense godoc
// @Summary      Update Expense
// @Description  updates an expense by id
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        id     path    string  true  "Expense ID"
// @Param        req    body    models.UpdateExpenseRequest true "Update Expense Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /expenses/{id} [put]
// @Security     ApiKeyAuth
func UpdateExpense(c *gin.Context) {
	var requestBody models.UpdateExpenseRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	expenseId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid expense id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.UpdateExpense(expenseId, userId.(primitive.ObjectID), requestBody)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Expense updated successfully"
	response.SendResponse(c)
}

// DeleteExpense godoc
// @Summary      Delete Expense
// @Description  deletes expense by id
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Expense ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /expenses/{id} [delete]
// @Security     ApiKeyAuth
func DeleteExpense(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	expenseId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid expense id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.DeleteExpense(expenseId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Expense deleted successfully"
	response.SendResponse(c)
}
