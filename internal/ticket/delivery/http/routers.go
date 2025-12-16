package http

import (
	"github.com/NamespaceManager/internal/middleware"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/gin-gonic/gin"
)

func MapTicketRoutes(ticketGroup *gin.RouterGroup, ticketHandler interfaces.TicketHandler) {
	// glidelet
	ticketGroup.POST("/updateTicketStatusFromGlidelet", ticketHandler.UpdateTicketStatusFromGlidelet())

	ticketGroup.Use(middleware.AuthMiddleware())
	ticketGroup.POST("/handleticket", ticketHandler.HandleTicketCallback())

	// ticket

	ticketGroup.POST("/requestTicketToCH", ticketHandler.RequestTicket())
	ticketGroup.GET("/cancelTicket/:ticket_id", ticketHandler.CancelTicket())
	ticketGroup.GET("/getUserTickets", ticketHandler.GetUserTickets())
	ticketGroup.GET("/getTickets/:namespace_id", ticketHandler.GetTicketByNamespaceID())
	// ticketGroup.GET("/getTicketFromCH/:namespace_id", ticketHandler.GetTicketFromCH())

	// task
	ticketGroup.POST("/useTickets", ticketHandler.UseTickets())        // create task
	ticketGroup.GET("/tasks", ticketHandler.GetTasks())                //get tasks
	ticketGroup.DELETE("/stopTask/:task_id", ticketHandler.StopTask()) // stop task

	// new
	ticketGroup.POST("/", ticketHandler.RequestTicket())
	ticketGroup.GET("/", ticketHandler.GetUserTickets())
	ticketGroup.PATCH("/:ticket_id/cancel", ticketHandler.CancelTicket())
	ticketGroup.GET("/namespace/:namespace_id", ticketHandler.GetTicketByNamespaceID())
}
