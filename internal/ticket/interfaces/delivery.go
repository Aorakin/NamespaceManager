package interfaces

import "github.com/gin-gonic/gin"

type TicketHandler interface {
	HandleTicketCallback() gin.HandlerFunc
	// SendTicket() gin.HandlerFunc
	GetTicketNS() gin.HandlerFunc
	RequestTicket() gin.HandlerFunc
	GetHistory() gin.HandlerFunc
	UseTickets() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
	GetTasks() gin.HandlerFunc
	RemoveTask() gin.HandlerFunc
	TicketStatus() gin.HandlerFunc
}
