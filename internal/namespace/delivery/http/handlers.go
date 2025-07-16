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

// Create godoc
// @Summary Create namespace
// @Description Create a new namespace
// @Tags namespaces
// @Accept json
// @Produce json
// @Param namespace body dtos.RequestNS true "Namespace data"
// @Success 200 {string} string "Namespace Created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "Conflict error"
// @Security ApiKeyAuth
// @Router /ns/nsCreate [post]
func (h NSHandlers) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.RequestNS
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		userID := utils.GetSession(c, "userID").(uuid.UUID)

		if err := h.nsUsecase.HandleCreate(req, userID); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, "Namespace Created")
	}
}

// GetNSList godoc
// @Summary Get namespace list
// @Description Get list of namespaces for the authenticated user
// @Tags namespaces
// @Produce json
// @Success 200 {object} map[string]interface{} "List of namespaces"
// @Failure 409 {object} map[string]string "Conflict error"
// @Security ApiKeyAuth
// @Router /ns/nsList [get]
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

func (h NSHandlers) AddUsersToNamespace() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.UpdateNamespaceUsersReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := h.nsUsecase.AddUsersToNamespace(req); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespace": "Users added to namespace successfully"})
	}
}

func (h NSHandlers) RemoveUsersFromNamespace() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.UpdateNamespaceUsersReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := h.nsUsecase.RemoveUsersFromNamespace(req); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespace": "Users removed from namespace successfully"})
	}
}

func (h NSHandlers) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.ReqForEdit
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		if err := h.nsUsecase.Update(req); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Update": "Success"})
	}
}

func (h NSHandlers) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.RequestNSID
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := h.nsUsecase.Delete(req.NamespaceID); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Delete": "Success"})
	}
}
