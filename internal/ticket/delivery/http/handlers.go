package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/NamespaceManager/internal/utils"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/NamespaceManager/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TicketHandlers struct {
	ticketUsecase interfaces.TicketUsecase
}

func NewTicketHandler(ticketUsecase interfaces.TicketUsecase) interfaces.TicketHandler {
	return &TicketHandlers{ticketUsecase: ticketUsecase}
}

func (h *TicketHandlers) HandleTicketCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketreq dtos.CreateTicket
		if err := c.ShouldBindJSON(&ticketreq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userid := utils.GetSession(c, "userID").(uuid.UUID)
		if err := h.ticketUsecase.HandleTicketCallback(ticketreq, userid); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, "Ticket Created")
	}
}

func (h *TicketHandlers) GetTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.MustGet("userID").(uuid.UUID)
		tasks, err := h.ticketUsecase.GetTasks(ownerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})

	}
}

func (h *TicketHandlers) StopTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		taskIDParam := c.Param("task_id")
		taskUUID, err := uuid.Parse(taskIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task_id"})
			return
		}

		response, err := h.ticketUsecase.StopTask(userID, taskUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func (h *TicketHandlers) UseTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketReq dtos.CreateTaskRequest
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userID := c.MustGet("userID").(uuid.UUID)

		// resp := make([][]map[string]interface{}, len(ticketReq.Tickets))
		// var successfulPayloads []models.GliderTicket
		// // var laststatus int
		// for i, payload := range ticketReq.Tickets {
		// 	_, jsonResponse, err := h.ticketUsecase.SendTicket(payload)
		// 	if err != nil {
		// 		if err := h.ticketUsecase.RollbackFailedTickets(successfulPayloads, i); err != nil {
		// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		// 		}
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 		return
		// 	}
		// 	successfulPayloads = append(successfulPayloads, payload)
		// 	resp[i] = jsonResponse
		// 	// laststatus = status
		// }
		fmt.Println(ticketReq.Tickets)
		// _, jsonResponse, err := h.ticketUsecase.SendTicket(ticketReq.Tickets)
		// if err != nil {
		// 	// if err := h.ticketUsecase.RollbackFailedTickets(successfulPayloads, i); err != nil {
		// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		// 	// }
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		if err := h.ticketUsecase.CreateTask(ticketReq, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, "Task Created")
	}
}

func (h *TicketHandlers) RequestTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		accessToken := c.MustGet("accessToken").(string)

		var request dtos.RequestTicketDTO
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		gliderTicket, err := h.ticketUsecase.RequestTicket(request, accessToken, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		// Format and return response
		c.JSON(http.StatusCreated, gin.H{"ticket": gliderTicket})
	}
}

func (h *TicketHandlers) GetTicketByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("namespace_id")
		log.Println("namespaceID", namespaceID)
		tickets, err := h.ticketUsecase.GetTicketByNamespaceID(namespaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tickets)
	}
}

func (h *TicketHandlers) GetUserTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		tickets, err := h.ticketUsecase.GetUserTickets(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tickets)
	}
}

func (h *TicketHandlers) GetTicketFromCH() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketId := c.Param("ticket_id")
		accessToken := c.MustGet("accessToken").(string)
		url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/" + url.PathEscape(ticketId)
		body, err := httpclient.SendRequestWithAccessToken(url, nil, "GET", accessToken)
		if err != nil {
			if apiErr, ok := err.(apiError.ApiErr); ok {
				c.JSON(apiErr.Status(), gin.H{"error": apiErr.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var ticketResponse dtos.GliderTicketResponse
		if err := json.Unmarshal(body, &ticketResponse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}
		c.JSON(http.StatusOK, ticketResponse)
	}
}

func (h *TicketHandlers) RequestTicketToCH() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		accessToken := c.MustGet("accessToken").(string)

		var ticketReq dtos.RequestTicketDTO
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/"
		res, err := httpclient.SendRequestWithAccessToken(url, ticketReq, "POST", accessToken)
		if err != nil {
			if apiErr, ok := err.(apiError.ApiErr); ok {
				c.JSON(apiErr.Status(), gin.H{"error": apiErr.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var gliderTicket dtos.GliderTicketResponse
		if err := json.Unmarshal(res, &gliderTicket); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}
		// status, gliderTicket, err := h.ticketUsecase.RequestTicketToCH(ticketReq)
		// fmt.Println("gliderTicket", gliderTicket)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }
		// // get ticket from ch เพื่อเอาไปเก็บ db
		// // status, ticketResponses, err := h.ticketUsecase.GetTicketFromCH(ticketdto.ID)
		// // if err != nil {
		// // 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// // 	return
		// // }
		// if status != http.StatusCreated {
		// 	c.JSON(status, gin.H{"error": "Failed to create ticket from Clearinghouse"})
		// 	return
		// }
		err = h.ticketUsecase.SaveTicket(gliderTicket, ticketReq.Name, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ticket": gliderTicket})

	}
}

func (h *TicketHandlers) UpdateTicketStatusFromGlidelet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []dtos.StatusRes
		log.Println(req)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		err := h.ticketUsecase.UpdateTicketStatusFromGlidelet(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Ticket statuses updated successfully"})
	}
}

func (h *TicketHandlers) CancelTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketId := c.Param("ticket_id")
		fmt.Println("ticketId", ticketId)
		if ticketId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_id is required"})
			return
		}
		url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/" + url.PathEscape(ticketId) + "/cancel"
		accessToken := c.MustGet("accessToken").(string)
		_, err := httpclient.SendRequestWithAccessToken(url, nil, "PATCH", accessToken)
		if err != nil {
			if apiErr, ok := err.(apiError.ApiErr); ok {
				c.JSON(apiErr.Status(), gin.H{"error": apiErr.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = h.ticketUsecase.CancelTicket(ticketId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Ticket cancelled successfully"})
	}
}
