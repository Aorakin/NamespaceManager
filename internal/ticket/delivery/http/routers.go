package http

import (
	"github.com/NamespaceManager/internal/middleware"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/gin-gonic/gin"
)

func MapTicketRoutes(ticketGroup *gin.RouterGroup, ticketHandler interfaces.TicketHandler) {
	ticketGroup.Use(middleware.AuthMiddleware())
	ticketGroup.POST("/handleticket", ticketHandler.HandleTicketCallback())
	// ticketGroup.POST("/sendticket", ticketHandler.SendTicket())
	// ticketGroup.POST("/getticket", ticketHandler.GetTicketNS())
	ticketGroup.POST("/requestTicket", ticketHandler.RequestTicket())
	// ticketGroup.GET("/history", ticketHandler.GetHistory())
	ticketGroup.POST("/useTickets", ticketHandler.UseTickets())
	// ticketGroup.PUT("/UpdateTicket", ticketHandler.Update())
	ticketGroup.GET("/tasks", ticketHandler.GetTasks())
	ticketGroup.DELETE("/stopTask", ticketHandler.StopTask())

	ticketGroup.GET("/getTickets/:namespace_id", ticketHandler.GetTicketByNamespaceID())
	ticketGroup.GET("/getUserTickets", ticketHandler.GetUserTickets())
	// ticketGroup.GET("/getTicketFromCH/:namespace_id", ticketHandler.GetTicketFromCH())
	ticketGroup.POST("/requestTicketToCH", ticketHandler.RequestTicketToCH())
	ticketGroup.GET("/cancelTicket/:ticket_id", ticketHandler.CancelTicket())

	ticketGroup.POST("/updateTicketStatusFromGlidelet", ticketHandler.UpdateTicketStatusFromGlidelet())
}
