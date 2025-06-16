package validators

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func SendFriendRequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sendFriendRequestRequest models.SendFriendRequestRequest
		_ = c.ShouldBindBodyWith(&sendFriendRequestRequest, binding.JSON)

		if err := sendFriendRequestRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}

func RespondFriendRequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var respondFriendRequestRequest models.RespondFriendRequestRequest
		_ = c.ShouldBindBodyWith(&respondFriendRequestRequest, binding.JSON)

		if err := respondFriendRequestRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}
