package http

import (
	"fmt"
	"net/http"

	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NSHandlers struct {
	nsUsecase interfaces.NSUsecase
}

func NewNSHandler(NSUsecase interfaces.NSUsecase) interfaces.NSHandler {
	return &NSHandlers{nsUsecase: NSUsecase}
}

func (h NSHandlers) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request dtos.RequestNS
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		userID := utils.GetSession(c, "userID").(uuid.UUID)

		if err := h.nsUsecase.HandleCreate(request, userID); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, "Namespace Created")
	}
}

func (h NSHandlers) GetNSList() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetSession(c, "userID").(uuid.UUID)
		fmt.Println(userID)
		NSList, err := h.nsUsecase.GetNSList(userID)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespace": NSList})
	}
}
