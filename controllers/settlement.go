package controllers

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetGroupBalances godoc
// @Summary      Get Group Balances
// @Description  gets balance summary for a group
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/balances [get]
// @Security     ApiKeyAuth
func GetGroupBalances(c *gin.Context) {
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

	balances, err := services.CalculateGroupBalances(groupId, userId.(primitive.ObjectID))
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
// @Description  calculates simplified settlement suggestions for a group
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/simplify [get]
// @Security     ApiKeyAuth
func SimplifyDebts(c *gin.Context) {
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

	settlements, err := services.SimplifyDebts(groupId, userId.(primitive.ObjectID))
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

// CreateSettlement godoc
// @Summary      Create Settlement
// @Description  creates a custom settlement between two users
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Param        req  body      models.CreateSettlementRequest true "Settlement Request"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /settlements [post]
// @Security     ApiKeyAuth
func CreateSettlement(c *gin.Context) {
	var requestBody models.CreateSettlementRequest
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

	// Only allow group members to create settlements
	_, err = services.GetGroupById(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = "access denied or group not found"
		response.SendResponse(c)
		return
	}

	settlement, err := services.CreateSettlement(groupId, payerId, payeeId, requestBody.Amount, requestBody.Currency, requestBody.Notes)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Success = true
	response.Data = gin.H{"settlement": settlement}
	response.SendResponse(c)
}

// MarkSettlementComplete godoc
// @Summary      Mark Settlement Complete
// @Description  marks a settlement as completed
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Settlement ID"
// @Param        req  body      models.SettleDebtRequest false "Settlement Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /settlements/{id}/complete [post]
// @Security     ApiKeyAuth
func MarkSettlementComplete(c *gin.Context) {
	var requestBody models.SettleDebtRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	settlementId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid settlement id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.MarkSettlementComplete(settlementId, userId.(primitive.ObjectID), requestBody.Notes)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Settlement marked as complete"
	response.SendResponse(c)
}

// GetUserSettlements godoc
// @Summary      Get User Settlements
// @Description  gets all settlements for the authenticated user
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Filter by status (pending, completed, cancelled)"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /settlements [get]
// @Security     ApiKeyAuth
func GetUserSettlements(c *gin.Context) {
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

	statusQuery := c.Query("status")
	var status db.SettlementStatus
	if statusQuery != "" {
		status = db.SettlementStatus(statusQuery)
	}

	settlements, err := services.GetUserSettlements(userId.(primitive.ObjectID), status)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"settlements": settlements}
	response.SendResponse(c)
}

// GetGroupSettlements godoc
// @Summary      Get Group Settlements
// @Description  gets all settlements for a specific group
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Param        groupId  path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{groupId}/settlements [get]
// @Security     ApiKeyAuth
func GetGroupSettlements(c *gin.Context) {
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

	settlements, err := services.GetGroupSettlements(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"settlements": settlements}
	response.SendResponse(c)
}
