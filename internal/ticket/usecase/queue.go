package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/google/uuid"
)

// getNextQueueTime retrieves the estimated start time of the next task in the queue.
// If there are no tasks in the queue, it returns a time far in the future.
func (u *TicketUsecase) getNextQueueTime(poolID uuid.UUID) (time.Time, error) {
	startTime, err := u.ticketRepository.GetNextStartTimeByPoolID(poolID)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get next queue task: %w", err)
	}
	if startTime.IsZero() {
		// No tasks in the queue, return a time far in the future
		return time.Now().Add(365 * 24 * time.Hour), nil
	}

	return startTime, nil
}

func (u *TicketUsecase) EnqueueTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	startTime, err := u.negotiateStartTime(ticketsByPool)
	log.Printf("[ENQUEUE TASK] Negotiated start time: %s", startTime.String())
	if err != nil {
		return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to negotiate start time: %w", err))
	}

	if startTime.IsZero() {
		return time.Time{}, nil
	}

	err = u.confirmTickets(ticketsByPool)
	if err != nil {
		return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to confirm tickets: %w", err))
	}
	log.Printf("[ENQUEUE TASK] Confirm ticket successfully")

	err = u.insertQueue(ticketsByPool, startTime)
	if err != nil {
		return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to insert tickets into queue: %w", err))
	}

	log.Printf("[ENQUEUE TASK] Insert queue to database, total pool: %d, start time %s", len(ticketsByPool), startTime.String())

	return startTime, nil
}

func (u *TicketUsecase) insertQueue(ticketsByPool map[uuid.UUID][]dtos.TicketReq, startTime time.Time) error {
	for poolID, poolTickets := range ticketsByPool {
		for _, ticket := range poolTickets {
			queueTicket := models.QueueTicket{
				GliderTicketID: ticket.ID,
				PoolID:         poolID,
				StartTime:      startTime,
			}

			err := u.ticketRepository.CreateQueueTicket(queueTicket)
			if err != nil {
				return fmt.Errorf("failed to insert queue ticket for pool %s: %w", poolID, err)
			}
		}
	}
	return nil
}

func (u *TicketUsecase) negotiateStartTime(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	var startTime time.Time

	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID.String(), poolTickets[0].GlideletURN)
		if err != nil {
			return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		nextQueueTime, err := u.getNextQueueTime(poolID)
		if err != nil {
			return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to get next queue time for pool %s: %w", poolID, err))
		}

		url := poolURN + "/api/v1/ticket/createList"
		payload := dtos.QueuePayload{
			Tickets: poolTickets,
			EndTime: nextQueueTime,
		}

		response, err := httpclient.SendRequest(url, payload, "POST")
		var poolResponse dtos.PoolQueueResponse
		if err := json.Unmarshal(response, &poolResponse); err != nil {
			return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to unmarshal response from pool %s: %w", poolID, err))
		}

		startTime = poolResponse.StartTime

		if err != nil {
			log.Println(string(response), err)
			return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to send tickets to pool %s: %w", poolID, err))
		}
	}

	return startTime, nil
}

// func (u *TicketUsecase) TryQueueTask() {
// 	tasks, err := u.ticketRepository.GetQueuedTasks()
// 	if err != nil {
// 		return
// 	}

// 	if len(tasks) == 0 {
// 		return
// 	}
// }
