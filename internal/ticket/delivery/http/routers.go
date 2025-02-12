package http

import (
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/gin-gonic/gin"
)

func MapTicketRoutes(ticketGroup *gin.RouterGroup, ticketHandler interfaces.TicketHandler) {
	ticketGroup.POST("/handleticket", ticketHandler.HandleTicketCallback())
	ticketGroup.POST("/sendticket", ticketHandler.SendTicket())
	ticketGroup.GET("/getticket", ticketHandler.GetMyTicket())
}
