package utils

import (
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/gin-gonic/gin"
	"net/http"
)

func TicketStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []dtos.StatusRes
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status received": req})

	}
}
