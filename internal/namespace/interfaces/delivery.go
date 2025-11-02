package interfaces

import "github.com/gin-gonic/gin"

type NSHandler interface {
	GetProjects() gin.HandlerFunc
	GetProjectDetail() gin.HandlerFunc
	GetNamespacesByProjectID() gin.HandlerFunc
	GetNamespacesDetail() gin.HandlerFunc
	GetQuotaByNamespaceID() gin.HandlerFunc
	GetProjectUsageByProjectID() gin.HandlerFunc
	GetNamespaceUsageByNamespaceID() gin.HandlerFunc
	GetQuotaUsageByNamespaceID() gin.HandlerFunc
	GetResource() gin.HandlerFunc
	GetResourcesPoolDetail() gin.HandlerFunc
}
