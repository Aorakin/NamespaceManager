package utils

import (
	"net/http"

	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/response"
	"github.com/gin-gonic/gin"
)

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
