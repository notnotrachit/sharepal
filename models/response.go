package models

import (
	"net/http"
	"time"

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

func SendSuccessResponse(c *gin.Context, message string, data map[string]any) {
	response := &Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    message,
		Data:       data,
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

// FriendRequestResponse represents a friend request with requester details
type FriendRequestResponse struct {
	ID             primitive.ObjectID `json:"id"`
	RequesterID    primitive.ObjectID `json:"requester_id"`
	RequesterName  string             `json:"requester_name"`
	RequesterEmail string             `json:"requester_email"`
	Status         string             `json:"status"`
	RequestedAt    time.Time          `json:"requested_at"`
}

// SentFriendRequestResponse represents a sent friend request with addressee details
type SentFriendRequestResponse struct {
	ID             primitive.ObjectID `json:"id"`
	AddresseeID    primitive.ObjectID `json:"addressee_id"`
	AddresseeName  string             `json:"addressee_name"`
	AddresseeEmail string             `json:"addressee_email"`
	Status         string             `json:"status"`
	RequestedAt    time.Time          `json:"requested_at"`
}
