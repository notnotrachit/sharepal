package validators

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func CreateExpenseValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createExpenseRequest models.CreateExpenseRequest
		_ = c.ShouldBindBodyWith(&createExpenseRequest, binding.JSON)

		if err := createExpenseRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}

func UpdateExpenseValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateExpenseRequest models.UpdateExpenseRequest
		_ = c.ShouldBindBodyWith(&updateExpenseRequest, binding.JSON)

		if err := updateExpenseRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}
