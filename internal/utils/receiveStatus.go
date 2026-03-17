package utils

import (
	"net/http"

	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/response"
	"github.com/gin-gonic/gin"
)

// TicketStatus godoc
// @Summary      Receive pod status payload
// @Description  Receive pod status updates posted to users podStatus endpoint
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        payload  body      []dtos.StatusRes  true  "Pod status payload"
// @Success      200      {object}  map[string]interface{}  "Status received"
// @Failure      400      {object}  response.ErrorResponse  "Bad request"
// @Router       /users/podStatus [post]
func TicketStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []dtos.StatusRes
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid input")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"status received": req})

	}
}
