package validators

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

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
