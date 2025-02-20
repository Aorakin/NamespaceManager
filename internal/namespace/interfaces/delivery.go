package interfaces

import "github.com/gin-gonic/gin"

type NSHandler interface {
	Create() gin.HandlerFunc
	GetNSList() gin.HandlerFunc
}
