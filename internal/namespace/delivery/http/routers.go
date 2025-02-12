package http

import (
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/gin-gonic/gin"
)

func MapNSRoutes(NSGroup *gin.RouterGroup, nsHandler interfaces.NSHandler) {
	NSGroup.POST("/nsCreate", nsHandler.Create())
}
