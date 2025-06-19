package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response Base response
type Response struct {
	StatusCode int            `json:"-"`
	Success    bool           `json:"success"`
	Message    string         `json:"message,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

func (response *Response) SendResponse(c *gin.Context) {
	c.AbortWithStatusJSON(response.StatusCode, response)
}

func SendResponseData(c *gin.Context, data gin.H) {
	response := &Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Data:       data,
	}
	response.SendResponse(c)
}

func SendErrorResponse(c *gin.Context, status int, message string) {
	response := &Response{
		StatusCode: status,
		Success:    false,
		Message:    message,
	}
	response.SendResponse(c)
}

// SettlementSuggestion represents a suggested settlement between two users
type SettlementSuggestion struct {
	GroupID   primitive.ObjectID `json:"group_id"`
	PayerID   primitive.ObjectID `json:"payer_id"`
	PayerName string             `json:"payer_name"`
	PayeeID   primitive.ObjectID `json:"payee_id"`
	PayeeName string             `json:"payee_name"`
	Amount    float64            `json:"amount"`
	Currency  string             `json:"currency"`
	Status    string             `json:"status"`
}
