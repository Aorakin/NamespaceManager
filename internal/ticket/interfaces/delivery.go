package interfaces

import "github.com/gin-gonic/gin"

type TicketHandler interface {
	HandleTicketCallback() gin.HandlerFunc
	SendTicket() gin.HandlerFunc
	GetMyTicket() gin.HandlerFunc
}
