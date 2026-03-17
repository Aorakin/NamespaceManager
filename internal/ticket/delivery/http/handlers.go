package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/NamespaceManager/internal/ticket/mapper"
	"github.com/NamespaceManager/internal/utils"
	apiError "github.com/NamespaceManager/pkg/api_error"
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

// HandleTicketCallback godoc
// @Summary      Handle ticket callback
// @Description  Create a ticket from callback payload
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        payload  body      dtos.CreateTicket  true  "Ticket callback payload"
// @Success      200      {object}  map[string]string  "Ticket created"
// @Failure      400      {object}  response.ErrorResponse  "Bad request"
// @Failure      401      {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/handleticket [post]
func (h *TicketHandlers) HandleTicketCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketreq dtos.CreateTicket
		if err := c.ShouldBindJSON(&ticketreq); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}
		userid := utils.GetSession(c, "userID").(uuid.UUID)
		if err := h.ticketUsecase.HandleTicketCallback(ticketreq, userid); err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Ticket created"})
	}
}

// GetTasks godoc
// @Summary      Get tasks
// @Description  Retrieve all tasks for the authenticated user
// @Tags         tasks
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Tasks list"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/tasks [get]
func (h *TicketHandlers) GetTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.MustGet("userID").(uuid.UUID)
		tasks, err := h.ticketUsecase.GetTasks(ownerID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}
}

// StopTask godoc
// @Summary      Stop task
// @Description  Stop a running task by ID
// @Tags         tasks
// @Produce      json
// @Param        task_id  path      string  true  "Task ID"
// @Success      200      {object}  map[string]interface{}  "Stop result"
// @Failure      400      {object}  response.ErrorResponse  "Bad request"
// @Failure      401      {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/stopTask/{task_id} [delete]
// @Router       /ticket/tasks/{task_id}/stop [patch]
func (h *TicketHandlers) StopTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		taskIDParam := c.Param("task_id")
		taskUUID, err := uuid.Parse(taskIDParam)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid task ID")))
			return
		}

		result, err := h.ticketUsecase.StopTask(userID, taskUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// UseTickets godoc
// @Summary      Create task from tickets
// @Description  Create a task using selected ticket IDs
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        payload  body      dtos.CreateTaskRequest  true  "Task creation payload"
// @Success      200      {object}  map[string]string       "Task created"
// @Failure      400      {object}  response.ErrorResponse  "Bad request"
// @Failure      401      {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/useTickets [post]
// @Router       /ticket/tasks [post]
func (h *TicketHandlers) UseTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var request dtos.CreateTaskRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}

		err := h.ticketUsecase.CreateTask(&request, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Task created"})
	}
}

// RequestTicket godoc
// @Summary      Request ticket
// @Description  Request a new ticket from clearing house
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        payload  body      dtos.RequestTicketDTO  true  "Ticket request payload"
// @Success      201      {object}  map[string]interface{}  "Created ticket"
// @Failure      400      {object}  response.ErrorResponse  "Bad request"
// @Failure      401      {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/requestTicketToCH [post]
// @Router       /ticket [post]
func (h *TicketHandlers) RequestTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		accessToken := c.MustGet("accessToken").(string)
		username := c.MustGet("firstname").(string) + " " + c.MustGet("lastname").(string)

		var request dtos.RequestTicketDTO
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid input")))
			return
		}

		gliderTicket, err := h.ticketUsecase.RequestTicket(request, accessToken, userID, username)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		// Format and return response
		c.JSON(http.StatusCreated, gin.H{"ticket": gliderTicket.Ticket})
	}
}

// GetTicketByNamespaceID godoc
// @Summary      Get tickets by namespace
// @Description  Retrieve all tickets belonging to a namespace
// @Tags         tickets
// @Produce      json
// @Param        namespace_id  path      string  true  "Namespace ID"
// @Success      200           {array}   dtos.UserTicketResponse
// @Failure      400           {object}  response.ErrorResponse  "Bad request"
// @Failure      500           {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/getTickets/{namespace_id} [get]
// @Router       /ticket/namespace/{namespace_id} [get]
func (h *TicketHandlers) GetTicketByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("namespace_id")

		namespaceUUID, err := uuid.Parse(namespaceID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid namespace ID")))
			return
		}

		tickets, err := h.ticketUsecase.GetTicketByNamespaceID(namespaceUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve tickets")))
			return
		}
		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

// GetTicketByNamespaceIDAndNodeID godoc
// @Summary      Get tickets by namespace and node
// @Description  Retrieve tickets in a namespace filtered by node ID
// @Tags         tickets
// @Produce      json
// @Param        namespace_id  path      string  true  "Namespace ID"
// @Param        node_id       path      string  true  "Node ID"
// @Success      200           {array}   dtos.UserTicketResponse
// @Failure      400           {object}  response.ErrorResponse  "Bad request"
// @Failure      500           {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/getTickets/{namespace_id}/{node_id} [get]
func (h *TicketHandlers) GetTicketByNamespaceIDAndNodeID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("namespace_id")
		nodeID := c.Param("node_id")

		namespaceUUID, err := uuid.Parse(namespaceID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid namespace ID")))
			return
		}

		nodeUUID, err := uuid.Parse(nodeID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid node ID")))
			return
		}

		tickets, err := h.ticketUsecase.GetTicketByNamespaceIDAndNodeID(namespaceUUID, nodeUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve tickets")))
			return
		}
		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

// GetUserTickets godoc
// @Summary      Get current user tickets
// @Description  Retrieve all tickets owned by the authenticated user
// @Tags         tickets
// @Produce      json
// @Success      200  {array}   dtos.UserTicketResponse
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/getUserTickets [get]
// @Router       /ticket [get]
func (h *TicketHandlers) GetUserTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		tickets, err := h.ticketUsecase.GetUserTickets(userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve tickets")))
			return
		}

		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

// UpdateTicketStatusFromGlidelet godoc
// @Summary      Update ticket status from glidelet
// @Description  Receive ticket status updates pushed from glidelet
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        payload  body      []dtos.StatusRes  true  "Status update payload"
// @Success      200      {object}  map[string]string  "Statuses updated"
// @Failure      400      {object}  map[string]string  "Bad request"
// @Failure      500      {object}  map[string]string  "Internal server error"
// @Router       /ticket/updateTicketStatusFromGlidelet [post]
func (h *TicketHandlers) UpdateTicketStatusFromGlidelet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []dtos.StatusRes
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if jsonData, err := json.MarshalIndent(req, "", "  "); err == nil {
			log.Println(string(jsonData))
		}

		err := h.ticketUsecase.UpdateTicketStatusFromGlidelet(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Ticket statuses updated successfully"})
	}
}

// CancelTicket godoc
// @Summary      Cancel ticket
// @Description  Cancel a ticket by ticket ID
// @Tags         tickets
// @Produce      json
// @Param        ticket_id  path      string  true  "Ticket ID"
// @Success      200        {object}  map[string]string       "Ticket cancelled"
// @Failure      400        {object}  response.ErrorResponse  "Bad request"
// @Failure      401        {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500        {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/cancelTicket/{ticket_id} [get]
// @Router       /ticket/{ticket_id}/cancel [patch]
func (h *TicketHandlers) CancelTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketId := c.Param("ticket_id")
		accessToken := c.MustGet("accessToken").(string)

		if ticketId == "" {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Ticket ID is required")))
			return
		}

		err := h.ticketUsecase.CancelTicket(ticketId, accessToken)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ticket cancelled successfully"})
	}
}

// CancelTask godoc
// @Summary      Cancel task
// @Description  Cancel a task by ID
// @Tags         tasks
// @Produce      json
// @Param        task_id  path      string  true  "Task ID"
// @Success      200      {object}  object  "No content body"
// @Failure      400      {object}  response.ErrorResponse  "Bad request"
// @Failure      401      {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/tasks/{task_id}/cancel [patch]
func (h *TicketHandlers) CancelTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		taskIDParam := c.Param("task_id")
		taskUUID, err := uuid.Parse(taskIDParam)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid task ID")))
			return
		}

		err = h.ticketUsecase.CancelTask(userID, taskUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

// DeleteTickets godoc
// @Summary      Delete tickets
// @Description  Soft-delete selected tickets by IDs
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        payload  body      dtos.DeleteTicketsRequest  true  "Ticket IDs"
// @Success      200      {object}  map[string]string          "Tickets deleted"
// @Failure      400      {object}  response.ErrorResponse     "Bad request"
// @Failure      401      {object}  response.ErrorResponse     "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse     "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/delete [patch]
func (h *TicketHandlers) DeleteTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		var req dtos.DeleteTicketsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}

		err := h.ticketUsecase.DeleteTickets(req, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "tickets deleted successfully"})
	}
}

// DeleteTasks godoc
// @Summary      Delete tasks
// @Description  Soft-delete selected tasks by IDs
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        payload  body      dtos.DeleteTasksRequest  true  "Task IDs"
// @Success      200      {object}  map[string]string        "Tasks deleted"
// @Failure      400      {object}  response.ErrorResponse   "Bad request"
// @Failure      401      {object}  response.ErrorResponse   "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse   "Internal server error"
// @Security     ApiKeyAuth
// @Router       /ticket/tasks/delete [patch]
func (h *TicketHandlers) DeleteTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		var req dtos.DeleteTasksRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}

		err := h.ticketUsecase.DeleteTasks(req, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "tasks deleted successfully"})
	}
}
