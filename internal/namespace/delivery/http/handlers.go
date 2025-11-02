package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/gin-gonic/gin"
)

type NSHandlers struct {
	nsUsecase interfaces.NSUsecase
}

func NewNSHandler(NSUsecase interfaces.NSUsecase) interfaces.NSHandler {
	return &NSHandlers{nsUsecase: NSUsecase}
}

func (h NSHandlers) GetProjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := os.Getenv("CLEARINGHOUSE_URL") + "/projects/all"

		status, body, err := utils.SendRequest(url, nil, "GET")
		fmt.Println(string(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var projects []dtos.ProjectDTO
		if err := json.Unmarshal(body, &projects); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse projects"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}
func (h NSHandlers) GetProjectDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/projects/" + projectID
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var project dtos.ProjectDTO
		if err := json.Unmarshal(body, &project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse project"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"project": project})
	}
}
func (h NSHandlers) GetNamespacesByProjectID() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/namespaces/all/" + projectID

		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var namespaces []dtos.NamespaceDTO
		if err := json.Unmarshal(body, &namespaces); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse namespaces"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
	}

}
func (h NSHandlers) GetNamespacesDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("ns_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/namespaces/" + namespaceID
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		fmt.Println(string(body))
		var namespace dtos.NamespaceDTO
		if err := json.Unmarshal(body, &namespace); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse namespace"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespace": namespace})
	}
}
func (h NSHandlers) GetQuotaByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("ns_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/quota/namespace/" + namespaceID
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var quotas []dtos.QuotaDTO
		if err := json.Unmarshal(body, &quotas); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse quotas"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"quotas": quotas})
	}

}
func (h NSHandlers) GetProjectUsageByProjectID() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/projects/" + projectID + "/usage"
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var usage dtos.UsageDTO
		if err := json.Unmarshal(body, &usage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse project usage"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"projectUsage": usage})
	}
}
func (h NSHandlers) GetNamespaceUsageByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("ns_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/namespaces/" + namespaceID + "/usage"
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var usage dtos.UsageDTO
		if err := json.Unmarshal(body, &usage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse namespace usage"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespaceUsage": usage})
	}
}
func (h NSHandlers) GetQuotaUsageByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		quotaID := c.Param("quota_id")
		namespaceID := c.Param("ns_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/quota/" + quotaID + "/usage/" + namespaceID
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var usage dtos.UsageDTO
		if err := json.Unmarshal(body, &usage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse quota usage"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"quotaUsage": usage})
	}
}
func (h NSHandlers) GetResource() gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID := c.Param("resource_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/resources/" + resourceID
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var resource dtos.ResourceDTO
		fmt.Println(string(body))
		if err := json.Unmarshal(body, &resource); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resource types"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resource": resource})
	}
}
func (h NSHandlers) GetResourcesPoolDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		poolID := c.Param("pool_id")
		url := os.Getenv("CLEARINGHOUSE_URL") + "/resources/pool/" + poolID
		status, body, err := utils.SendRequest(url, nil, "GET")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if status != http.StatusOK {
			c.JSON(status, gin.H{"error": string(body)})
			return
		}
		var pool dtos.ResourcesPoolDetailDTO
		if err := json.Unmarshal(body, &pool); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resource pool"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resourcePool": pool})
	}
}
