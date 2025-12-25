package interfaces

import "github.com/gin-gonic/gin"

type TicketHandler interface {
	HandleTicketCallback() gin.HandlerFunc
	UseTickets() gin.HandlerFunc
	GetTasks() gin.HandlerFunc
	StopTask() gin.HandlerFunc
	GetUserTickets() gin.HandlerFunc
	GetTicketByNamespaceID() gin.HandlerFunc
	UpdateTicketStatusFromGlidelet() gin.HandlerFunc
	CancelTicket() gin.HandlerFunc
	RequestTicket() gin.HandlerFunc
	CancelTask() gin.HandlerFunc
}
