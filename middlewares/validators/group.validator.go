package validators

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func CreateGroupValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createGroupRequest models.CreateGroupRequest
		_ = c.ShouldBindBodyWith(&createGroupRequest, binding.JSON)

		if err := createGroupRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}

func AddMemberToGroupValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var addMemberRequest models.AddMemberToGroupRequest
		_ = c.ShouldBindBodyWith(&addMemberRequest, binding.JSON)

		if err := addMemberRequest.Validate(); err != nil {
			models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Next()
	}
}
