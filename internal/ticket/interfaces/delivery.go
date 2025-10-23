package interfaces

import "github.com/gin-gonic/gin"

type TicketHandler interface {
	HandleTicketCallback() gin.HandlerFunc
	// SendTicket() gin.HandlerFunc
	// GetTicketNS() gin.HandlerFunc
	RequestTicket() gin.HandlerFunc
	// GetHistory() gin.HandlerFunc
	UseTickets() gin.HandlerFunc
	Delete() gin.HandlerFunc
	// Update() gin.HandlerFunc
	GetTasks() gin.HandlerFunc
	StopTask() gin.HandlerFunc
	GetTicketFromCH() gin.HandlerFunc
	GetUserTickets() gin.HandlerFunc
	RequestTicketToCH() gin.HandlerFunc
	GetTicketByNamespaceID() gin.HandlerFunc
	UpdateTicketStatusFromGlidelet() gin.HandlerFunc
	CancelTicket() gin.HandlerFunc
}
