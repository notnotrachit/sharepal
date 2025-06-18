package validators

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func CreateSettlementValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createSettlementRequest models.CreateSettlementRequest
		_ = c.ShouldBindBodyWith(&createSettlementRequest, binding.JSON)

		if err := createSettlementRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}

func SettleDebtValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var settleDebtRequest models.SettleDebtRequest
		_ = c.ShouldBindBodyWith(&settleDebtRequest, binding.JSON)

		if err := settleDebtRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}
