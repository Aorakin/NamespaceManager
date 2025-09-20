package http

import (
	"github.com/NamespaceManager/internal/middleware"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/gin-gonic/gin"
)

func MapNSRoutes(NSGroup *gin.RouterGroup, nsHandler interfaces.NSHandler) {
	NSGroup.Use(middleware.AuthMiddleware())
	NSGroup.GET("/projects", nsHandler.GetProjects())
	NSGroup.GET("/projects/:project_id", nsHandler.GetProjectDetail())
	NSGroup.GET("/projectUsage/:project_id", nsHandler.GetProjectUsageByProjectID())

	NSGroup.GET("/namespaces/all/:project_id", nsHandler.GetNamespacesByProjectID())
	NSGroup.GET("/namespaces/:ns_id", nsHandler.GetNamespacesDetail())
	NSGroup.GET("/namespaceUsage/:ns_id", nsHandler.GetNamespaceUsageByNamespaceID())

	NSGroup.GET("/quota/:ns_id", nsHandler.GetQuotaByNamespaceID())
}
