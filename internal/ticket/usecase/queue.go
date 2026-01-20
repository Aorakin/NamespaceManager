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
func (u *TicketUsecase) getNextQueueTime(poolID uuid.UUID, NodeNames []string) (time.Time, error) {
	return time.Now().Add(365 * 24 * time.Hour), nil // Default to far future
	// startTime, err := u.ticketRepository.GetNextStartTime(poolID, NodeNames)
	// if err != nil {
	// 	return time.Time{}, fmt.Errorf("failed to get next queue task: %w", err)
	// }
	// if startTime.IsZero() {
	// 	// No tasks in the queue, return a time far in the future
	// 	return time.Now().Add(365 * 24 * time.Hour), nil
	// }

	// return startTime, nil
}

func (u *TicketUsecase) allNodesHaveHeadTasks(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (bool, error) {
	for poolID, poolTickets := range ticketsByPool {
		nodeNames := u.getNodeNames(poolTickets)
		headTasks, err := u.ticketRepository.GetHeadTasksByPoolAndNodes(poolID, nodeNames)
		if err != nil {
			return false, fmt.Errorf("failed to get head tasks for pool %s: %w", poolID, err)
		}

		for _, nodeName := range nodeNames {
			if headTasks[nodeName] == nil {
				return false, nil
			}
		}
	}
	return true, nil
}

func (u *TicketUsecase) backfillTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	startTime, err := u.tryBackfillTask(ticketsByPool)
	log.Printf("[ENQUEUE TASK] Negotiated start time: %s", startTime.String())
	err = u.confirmTickets(ticketsByPool)
	if err != nil {
		log.Printf("[ENQUEUE TASK] Confirm tickets return with ERROR : %s", err.Error())
		// return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to confirm tickets: %s", err.Error()))
	}
	return startTime, nil
}

func (u *TicketUsecase) tryBackfillTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	backfillSuccess := true
	startTime := time.Now()

	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID.String(), poolTickets[0].GlideletURN)
		log.Printf("[BACKFILL TASK] PoolID: %s, PoolURN: %s", poolID, poolURN)
		if err != nil {
			return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		url := poolURN + "/api/v1/ticket/backFillJobs"
		payload := dtos.QueuePayload{
			Tickets:   poolTickets,
			StartTime: startTime,
		}

		response, err := httpclient.SendRequest(url, payload, "POST")
		if err != nil {
			log.Printf("[BACKFILL TASK] PoolID: %s, Error: %v", poolID, err.Error())
			backfillSuccess = false
			break
		}
		log.Printf("[BACKFILL TASK] PoolID: %s, Response: %s", poolID, string(response))

		var poolResponse dtos.PoolBackfillResponse
		if err := json.Unmarshal(response, &poolResponse); err != nil {
			log.Printf("[BACKFILL TASK] Failed to unmarshal response from pool %s: %v", poolID, err)
			backfillSuccess = false
			break
		}

		if !poolResponse.Status {
			log.Printf("[BACKFILL TASK] PoolID: %s, Backfill failed (status: false)", poolID)
			backfillSuccess = false
			break
		}
	}

	if !backfillSuccess {
		log.Printf("[BACKFILL TASK] Backfill failed, falling back to cancel tickets")
		if err := u.fallBackQueueTask(ticketsByPool); err != nil {
			return time.Time{}, err
		}
		return time.Time{}, nil
	}

	log.Printf("[BACKFILL TASK] Backfill successful for all pools")
	return startTime, nil
}

func (u *TicketUsecase) fallBackQueueTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) error {
	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID.String(), poolTickets[0].GlideletURN)
		log.Printf("[FALLBACK QUEUE TASK] PoolID: %s, PoolURN: %s", poolID, poolURN)
		if err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		var ticketIDs []uuid.UUID
		for _, ticket := range poolTickets {
			ticketIDs = append(ticketIDs, ticket.ID)
		}

		payload := dtos.StopTaskTickets{TicketIDs: ticketIDs}
		url := poolURN + "/api/v1/ticket/cancelPods"

		_, err = httpclient.SendRequest(url, payload, "PATCH")
		if err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to cancel tickets in pool %s: %w", poolID, err))
		}
	}
	return nil
}

func (u *TicketUsecase) queueHeadTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	startTime, err := u.negotiateStartTime(ticketsByPool)
	log.Printf("[ENQUEUE TASK] Negotiated start time: %s", startTime.String())
	if err != nil {
		log.Printf("[ENQUEUE TASK] Negotiated return with ERROR : %s", err.Error())
		return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to negotiate start time: %s", err.Error()))
	}

	log.Printf("[ENQUEUE TASK] Negotiated start time: %s", startTime.String())
	err = u.confirmTickets(ticketsByPool)
	if err != nil {
		log.Printf("[ENQUEUE TASK] Confirm tickets return with ERROR : %s", err.Error())
		// return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to confirm tickets: %s", err.Error()))
	}

	err = u.insertQueue(ticketsByPool, startTime)
	if err != nil {
		return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to insert tickets into queue: %s", err.Error()))
	}

	log.Printf("[ENQUEUE TASK] Insert queue to database, total pool: %d, start time %s", len(ticketsByPool), startTime.String())

	return startTime, nil
}

// EnqueueTask handles the logic of enqueuing tasks based on current queue status.
// Requirement is guarantees that all tickets start at the same time.
//
// Example Scenario:
//
//	Job A requests tickets for Node1 and Node2 for 2 hours
//	Job B requests tickets for Node2 and Node3 for 2 hours
//	Job C requests tickets for Node3 for 10 hours
//
// Design Approaches:
//
// Global Head Task:
//   - Pros: Simpler to manage (only one head task)
//   - Cons: Job C blocks Node3 for 10 hours, preventing Job B from using it
//
// Multiple Head Task:
//   - Each node can have its own head task, but same start time is enforced across tickets.
//   - Pros: Job B can guarantee to start after Job A which make Job C being pushed to wait after Job B finish instead of delayed Job B
//   - Cons: More complex to manage multiple head tasks
//
// Final Decision:
// We use the Multiple Head Task approach to reduce wait times.
//
// Algorithm:
//
//  1. Check if all nodes in ticketsByPool have an existing head task
//  2. If yes: Try to backfill the earliest head task that can accommodate all nodes
//     → Call tryBackFillTask(ticketsByPool)
//  3. If no: Create a new head task for the ticketsByPool
//     (Some nodes may have existing head tasks; this becomes their second head task)
//     → Call queueHeadTask(ticketsByPool)
func (u *TicketUsecase) EnqueueTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	isTrue, err := u.allNodesHaveHeadTasks(ticketsByPool)
	log.Printf("[ENQUEUE TASK] all node have head tasks: %v", isTrue)
	if err != nil {
		return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to get head task: %s", err.Error()))
	}

	if !isTrue {
		return u.queueHeadTask(ticketsByPool)
	}

	return u.backfillTask(ticketsByPool)

}

func (u *TicketUsecase) insertQueue(ticketsByPool map[uuid.UUID][]dtos.TicketReq, startTime time.Time) error {
	for poolID, poolTickets := range ticketsByPool {
		for _, ticket := range poolTickets {
			queueTicket := models.QueueTicket{
				GliderTicketID: ticket.ID,
				PoolID:         poolID,
				NodeName:       ticket.NodeName,
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

func (u *TicketUsecase) getNodeNames(tickets []dtos.TicketReq) []string {
	var nodeNames []string
	for _, ticket := range tickets {
		nodeNames = append(nodeNames, ticket.NodeName)
	}
	return nodeNames
}

func (u *TicketUsecase) negotiateStartTime(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
	startTime := time.Now()
	maxIterations := 20
	iteration := 0

	log.Printf("[NEGOTIATE START TIME] Starting multi-pool negotiation with %d pool(s)", len(ticketsByPool))

	for iteration < maxIterations {
		iteration++
		log.Printf("[NEGOTIATE START TIME] Iteration %d, trying with start time: %s", iteration, startTime.String())

		poolResponses := make(map[uuid.UUID]time.Time)
		hasError := false

		// Ask each pool for a slot starting from current startTime
		for poolID, poolTickets := range ticketsByPool {
			poolURN, err := u.getPoolURN(poolID.String(), poolTickets[0].GlideletURN)
			if err != nil {
				return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
			}

			url := poolURN + "/api/v1/ticket/createList"
			payload := dtos.QueuePayload{
				Tickets:   poolTickets,
				StartTime: startTime,
			}

			response, err := httpclient.SendRequest(url, payload, "POST")
			if err != nil {
				log.Printf("[NEGOTIATE START TIME] PoolID: %s, Error: %v", poolID, err.Error())
				hasError = true
				break
			}

			var poolResponse dtos.PoolQueueResponse
			if err := json.Unmarshal(response, &poolResponse); err != nil {
				log.Printf("[NEGOTIATE START TIME] Failed to unmarshal response from pool %s: %v", poolID, err)
				hasError = true
				break
			}

			poolResponses[poolID] = poolResponse.StartTime
			log.Printf("[NEGOTIATE START TIME] PoolID: %s returned start time: %s", poolID, poolResponse.StartTime.String())
		}

		if hasError {
			log.Printf("[NEGOTIATE START TIME] Error occurred, calling fallback and retrying")
			if err := u.fallBackQueueTask(ticketsByPool); err != nil {
				log.Printf("[NEGOTIATE START TIME] Fallback failed: %v", err)
			}
			startTime = startTime.Add(time.Minute)
			continue
		}

		// Check if any pool returned zero time (resources not available)
		hasZeroTime := false
		for poolID, t := range poolResponses {
			if t.IsZero() {
				log.Printf("[NEGOTIATE START TIME] Pool %s returned zero time (resources not available)", poolID)
				hasZeroTime = true
				break
			}
		}

		if hasZeroTime {
			log.Printf("[NEGOTIATE START TIME] Pool returned zero time, calling fallback and retrying")
			if err := u.fallBackQueueTask(ticketsByPool); err != nil {
				log.Printf("[NEGOTIATE START TIME] Fallback failed: %v", err)
			}
			startTime = startTime.Add(time.Minute)
			continue
		}

		// Check if all pools returned the same start time
		isAllSame := true
		var firstTime time.Time
		firstPoolID := ""

		for poolID, t := range poolResponses {
			if firstTime.IsZero() {
				firstTime = t
				firstPoolID = poolID.String()
			} else if !t.Equal(firstTime) {
				isAllSame = false
				log.Printf("[NEGOTIATE START TIME] Time mismatch: Pool %s (%s) != Pool %s (%s)",
					firstPoolID, firstTime.String(), poolID, t.String())
				break
			}
		}

		// If all pools agreed on the same time, we're done
		if isAllSame {
			log.Printf("[NEGOTIATE START TIME] All pools agreed on start time: %s after %d iteration(s)", firstTime.String(), iteration)
			return firstTime, nil
		}

		// Pools didn't agree, call fallback before next iteration
		log.Printf("[NEGOTIATE START TIME] Pools didn't agree on time, calling fallback before retry")
		if err := u.fallBackQueueTask(ticketsByPool); err != nil {
			log.Printf("[NEGOTIATE START TIME] Fallback failed: %v", err)
		}

		// Find the latest time among all pool responses
		latestTime := startTime
		for _, t := range poolResponses {
			if t.After(latestTime) {
				latestTime = t
			}
		}

		// Move to the latest time for next iteration
		startTime = latestTime
		log.Printf("[NEGOTIATE START TIME] Moving to latest time: %s for next iteration", latestTime.String())
	}

	// Max iterations reached, call fallback and return error
	log.Printf("[NEGOTIATE START TIME] Failed to negotiate after %d iterations, calling fallback", maxIterations)
	if err := u.fallBackQueueTask(ticketsByPool); err != nil {
		log.Printf("[NEGOTIATE START TIME] Final fallback failed: %v", err)
	}
	return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to negotiate start time after %d iterations", maxIterations))
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
