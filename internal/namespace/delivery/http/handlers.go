package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/NamespaceManager/internal/utils"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/NamespaceManager/pkg/response"
	"github.com/gin-gonic/gin"
)

type NSHandlers struct {
	nsUsecase interfaces.NSUsecase
}

func NewNSHandler(NSUsecase interfaces.NSUsecase) interfaces.NSHandler {
	return &NSHandlers{nsUsecase: NSUsecase}
}

// GetProjects godoc
// @Summary      Get all projects
// @Description  Retrieve a list of all projects from the clearing house
// @Tags         namespaces
// @Produce      json
// @Success      200  {object}  map[string][]dtos.ProjectDTO  "List of projects"
// @Failure      500  {object}  response.ErrorResponse         "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/projects [get]
func (h NSHandlers) GetProjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := os.Getenv("CLEARINGHOUSE_URL") + "/projects/all"
		accessToken := c.MustGet("accessToken").(string)

		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		fmt.Println(string(body))
		if err != nil {
			log.Printf("failed to fetch projects: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load projects, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load projects")))
			return
		}
		var projects []dtos.ProjectDTO
		if err := json.Unmarshal(body, &projects); err != nil {
			log.Printf("failed to parse projects response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process projects data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

// GetProjectDetail godoc
// @Summary      Get project detail
// @Description  Retrieve details of a specific project by ID
// @Tags         namespaces
// @Produce      json
// @Param        project_id  path      string  true  "Project ID"
// @Success      200         {object}  map[string]dtos.ProjectDTO  "Project details"
// @Failure      500         {object}  response.ErrorResponse       "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/projects/{project_id} [get]
func (h NSHandlers) GetProjectDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/projects/" + projectID
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch project detail: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load project details, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load project details")))
			return
		}
		var project dtos.ProjectDTO
		if err := json.Unmarshal(body, &project); err != nil {
			log.Printf("failed to parse project response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process project data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"project": project})
	}
}

// GetNamespacesByProjectID godoc
// @Summary      Get namespaces by project ID
// @Description  Retrieve all namespaces belonging to a specific project
// @Tags         namespaces
// @Produce      json
// @Param        project_id  path      string  true  "Project ID"
// @Success      200         {object}  map[string][]dtos.NamespaceDTO  "List of namespaces"
// @Failure      500         {object}  response.ErrorResponse           "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/namespaces/all/{project_id} [get]
func (h NSHandlers) GetNamespacesByProjectID() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/namespaces/all/" + projectID

		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch namespaces: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load namespaces, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load namespaces")))
			return
		}
		var namespaces []dtos.NamespaceDTO
		if err := json.Unmarshal(body, &namespaces); err != nil {
			log.Printf("failed to parse namespaces response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process namespaces data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
	}

}

// GetNamespacesDetail godoc
// @Summary      Get namespace detail
// @Description  Retrieve details of a specific namespace by ID
// @Tags         namespaces
// @Produce      json
// @Param        ns_id  path      string  true  "Namespace ID"
// @Success      200    {object}  map[string]dtos.NamespaceDTO  "Namespace details"
// @Failure      500    {object}  response.ErrorResponse         "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/namespaces/{ns_id} [get]
func (h NSHandlers) GetNamespacesDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("ns_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/namespaces/" + namespaceID
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch namespace detail: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load namespace details, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load namespace details")))
			return
		}
		fmt.Println(string(body))
		var namespace dtos.NamespaceDTO
		if err := json.Unmarshal(body, &namespace); err != nil {
			log.Printf("failed to parse namespace response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process namespace data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespace": namespace})
	}
}

// GetQuotaByNamespaceID godoc
// @Summary      Get quota by namespace ID
// @Description  Retrieve quota information for a specific namespace
// @Tags         namespaces
// @Produce      json
// @Param        ns_id  path      string  true  "Namespace ID"
// @Success      200    {object}  object  "Quota data (JSON from clearing house)"
// @Failure      400    {object}  response.ErrorResponse  "Bad request"
// @Failure      500    {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/quota/{ns_id} [get]
func (h NSHandlers) GetQuotaByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("ns_id")
		accessToken := c.MustGet("accessToken").(string)

		if namespaceID == "" {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Namespace ID is required")))
			return
		}

		url := os.Getenv("CLEARINGHOUSE_URL") + "/quota/namespace/" + namespaceID
		quotaData, err := httpclient.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch quota: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to load quota information")))
			return
		}

		c.Data(http.StatusOK, "application/json", quotaData)
	}

}

// GetProjectUsageByProjectID godoc
// @Summary      Get project usage
// @Description  Retrieve usage statistics for a specific project
// @Tags         namespaces
// @Produce      json
// @Param        project_id  path      string  true  "Project ID"
// @Success      200         {object}  map[string]dtos.UsageDTO  "Project usage data"
// @Failure      500         {object}  response.ErrorResponse     "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/projectUsage/{project_id} [get]
func (h NSHandlers) GetProjectUsageByProjectID() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/projects/" + projectID + "/usage"
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch project usage: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load project usage, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load project usage")))
			return
		}
		var usage dtos.UsageDTO
		if err := json.Unmarshal(body, &usage); err != nil {
			log.Printf("failed to parse project usage response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process project usage data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"projectUsage": usage})
	}
}

// GetNamespaceUsageByNamespaceID godoc
// @Summary      Get namespace usage
// @Description  Retrieve usage statistics for a specific namespace
// @Tags         namespaces
// @Produce      json
// @Param        ns_id  path      string  true  "Namespace ID"
// @Success      200    {object}  map[string]dtos.UsageDTO  "Namespace usage data"
// @Failure      500    {object}  response.ErrorResponse     "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/namespaceUsage/{ns_id} [get]
func (h NSHandlers) GetNamespaceUsageByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("ns_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/namespaces/" + namespaceID + "/usage"
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch namespace usage: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load namespace usage, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load namespace usage")))
			return
		}
		var usage dtos.UsageDTO
		if err := json.Unmarshal(body, &usage); err != nil {
			log.Printf("failed to parse namespace usage response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process namespace usage data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"namespaceUsage": usage})
	}
}

// GetQuotaUsageByNamespaceID godoc
// @Summary      Get quota usage by namespace ID
// @Description  Retrieve quota usage statistics for a specific quota in a namespace
// @Tags         namespaces
// @Produce      json
// @Param        quota_id  path      string  true  "Quota ID"
// @Param        ns_id     path      string  true  "Namespace ID"
// @Success      200       {object}  map[string]dtos.UsageDTO  "Quota usage data"
// @Failure      500       {object}  response.ErrorResponse     "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/quotaUsage/{quota_id}/{ns_id} [get]
func (h NSHandlers) GetQuotaUsageByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		quotaID := c.Param("quota_id")
		namespaceID := c.Param("ns_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/quota/" + quotaID + "/usage/" + namespaceID
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch quota usage: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load quota usage, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load quota usage")))
			return
		}
		var usage dtos.UsageDTO
		if err := json.Unmarshal(body, &usage); err != nil {
			log.Printf("failed to parse quota usage response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process quota usage data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"quotaUsage": usage})
	}
}

// GetResource godoc
// @Summary      Get resource detail
// @Description  Retrieve details of a specific resource by ID
// @Tags         namespaces
// @Produce      json
// @Param        resource_id  path      string  true  "Resource ID"
// @Success      200          {object}  map[string]interface{}  "Resource details"
// @Failure      500          {object}  response.ErrorResponse        "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/resource/{resource_id} [get]
func (h NSHandlers) GetResource() gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID := c.Param("resource_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/resources/" + resourceID
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch resource: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load resource, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load resource")))
			return
		}
		var resource dtos.ResourceDTO
		fmt.Println(string(body))
		if err := json.Unmarshal(body, &resource); err != nil {
			log.Printf("failed to parse resource response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process resource data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"resource": resource})
	}
}

// GetResourcesPoolDetail godoc
// @Summary      Get resource pool detail
// @Description  Retrieve resource pool details by pool ID
// @Tags         namespaces
// @Produce      json
// @Param        pool_id  path      string  true  "Pool ID"
// @Success      200      {object}  map[string]dtos.ResourcesPoolDetailDTO  "Resource pool details"
// @Failure      500      {object}  response.ErrorResponse                   "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ns/pool/{pool_id} [get]
func (h NSHandlers) GetResourcesPoolDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		poolID := c.Param("pool_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/resources/node/" + poolID
		status, body, err := utils.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			log.Printf("failed to fetch resource pool: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Unable to load resource pool, please try again later")))
			return
		}
		if status != http.StatusOK {
			log.Printf("external service returned status %d: %s", status, string(body))
			c.JSON(response.ErrorResponseBuilder(apiError.NewApiError(status, "Request failed", "Failed to load resource pool")))
			return
		}
		var pool dtos.ResourcesPoolDetailDTO
		if err := json.Unmarshal(body, &pool); err != nil {
			log.Printf("failed to parse resource pool response: %v", err)
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to process resource pool data")))
			return
		}
		c.JSON(http.StatusOK, gin.H{"resourcePool": pool})
	}
}
