package interfaces

import "github.com/gin-gonic/gin"

type NSHandler interface {
	Create() gin.HandlerFunc
	GetNSList() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
	RemoveUsersFromNamespace() gin.HandlerFunc
	AddUsersToNamespace() gin.HandlerFunc
}
