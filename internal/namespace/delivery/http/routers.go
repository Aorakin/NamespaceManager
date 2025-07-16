package http

import (
	"github.com/NamespaceManager/internal/middleware"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/gin-gonic/gin"
)

func MapNSRoutes(NSGroup *gin.RouterGroup, nsHandler interfaces.NSHandler) {
	NSGroup.Use(middleware.AuthMiddleware())
	NSGroup.POST("/nsCreate", nsHandler.Create())
	NSGroup.GET("/nsList", nsHandler.GetNSList())
	NSGroup.PUT("/nsUpdate", nsHandler.Update())
	NSGroup.DELETE("/nsDelete", nsHandler.Delete())
	NSGroup.PUT("/nsAddUsers", nsHandler.AddUsersToNamespace())
	NSGroup.PUT("/nsRemoveUsers", nsHandler.RemoveUsersFromNamespace())
}
