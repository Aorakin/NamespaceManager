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
	u.enqueueMu.Lock()
	defer u.enqueueMu.Unlock()
	return u.enqueueTask(ticketsByPool)
}

// enqueueTask is the internal implementation, assumes lock is already held.
func (u *TicketUsecase) enqueueTask(ticketsByPool map[uuid.UUID][]dtos.TicketReq) (time.Time, error) {
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

	log.Printf("[NEGOTIATE START TIME] Starting multi-pool negotiation with %d pool(s)", len(ticketsByPool))

	for iteration := 1; iteration <= maxIterations; iteration++ {
		log.Printf("[NEGOTIATE START TIME] Iteration %d, trying with start time: %s", iteration, startTime.String())

		type poolResult struct {
			poolID    uuid.UUID
			startTime time.Time
			err       error
		}

		results := make(chan poolResult, len(ticketsByPool))
		for poolID, poolTickets := range ticketsByPool {
			go func(poolID uuid.UUID, poolTickets []dtos.TicketReq) {
				poolURN, err := u.getPoolURN(poolID.String(), poolTickets[0].GlideletURN)
				if err != nil {
					results <- poolResult{poolID: poolID, err: fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err)}
					return
				}

				payload := dtos.QueuePayload{Tickets: poolTickets, StartTime: startTime}
				response, err := httpclient.SendRequest(poolURN+"/api/v1/ticket/createList", payload, "POST")
				if err != nil {
					results <- poolResult{poolID: poolID, err: err}
					return
				}

				var poolResponse dtos.PoolQueueResponse
				if err := json.Unmarshal(response, &poolResponse); err != nil {
					results <- poolResult{poolID: poolID, err: err}
					return
				}

				results <- poolResult{poolID: poolID, startTime: poolResponse.StartTime}
			}(poolID, poolTickets)
		}

		// Collect results
		poolResponses := make(map[uuid.UUID]time.Time)
		hasError := false
		for range ticketsByPool {
			res := <-results
			if res.err != nil {
				log.Printf("[NEGOTIATE START TIME] PoolID: %s, Error: %v", res.poolID, res.err)
				hasError = true
				continue
			}
			if res.startTime.IsZero() {
				log.Printf("[NEGOTIATE START TIME] Pool %s returned zero time", res.poolID)
				hasError = true
				continue
			}
			poolResponses[res.poolID] = res.startTime
			log.Printf("[NEGOTIATE START TIME] PoolID: %s returned start time: %s", res.poolID, res.startTime.String())
		}

		if hasError {
			log.Printf("[NEGOTIATE START TIME] Error occurred, calling fallback and retrying")
			if err := u.fallBackQueueTask(ticketsByPool); err != nil {
				log.Printf("[NEGOTIATE START TIME] Fallback failed: %v", err)
			}
			startTime = startTime.Add(time.Minute)
			continue
		}

		// Find latest time and check if all pools agree
		latestTime := startTime
		isAllSame := true
		var firstTime time.Time

		for poolID, t := range poolResponses {
			if firstTime.IsZero() {
				firstTime = t
			} else if !t.Equal(firstTime) {
				isAllSame = false
				log.Printf("[NEGOTIATE START TIME] Time mismatch at pool %s: %s != %s", poolID, t.String(), firstTime.String())
			}
			if t.After(latestTime) {
				latestTime = t
			}
		}

		if isAllSame {
			log.Printf("[NEGOTIATE START TIME] All pools agreed on start time: %s after %d iteration(s)", firstTime.String(), iteration)
			return firstTime, nil
		}

		log.Printf("[NEGOTIATE START TIME] Pools didn't agree, calling fallback and moving to latest time: %s", latestTime.String())
		if err := u.fallBackQueueTask(ticketsByPool); err != nil {
			log.Printf("[NEGOTIATE START TIME] Fallback failed: %v", err)
		}
		startTime = latestTime
	}

	log.Printf("[NEGOTIATE START TIME] Failed to negotiate after %d iterations, calling fallback", maxIterations)
	if err := u.fallBackQueueTask(ticketsByPool); err != nil {
		log.Printf("[NEGOTIATE START TIME] Final fallback failed: %v", err)
	}
	return time.Time{}, apiError.NewInternalServerError(fmt.Errorf("failed to negotiate start time after %d iterations", maxIterations))
}

// requeue handles re-enqueuing tasks associated with freed nodes.
// Called when a resource is finished, stopped, or expired on a node.
// It clears all queue entries for the affected nodes, then re-enqueues
// the tasks in FIFO order (by created_at) so they can negotiate earlier start times
// against the newly freed resources.
func (u *TicketUsecase) requeue(nodeNames []string) error {
	u.enqueueMu.Lock()
	defer u.enqueueMu.Unlock()

	queuedTasks, err := u.ticketRepository.GetTasksByNodeNames(nodeNames)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to get tasks by node name: %w", err))
	}
	log.Printf("[REQUEUE] Found %d tasks for nodes %v", len(queuedTasks), nodeNames)

	// Pre-compute formatted tickets for all tasks
	formattedTickets := make(map[uuid.UUID]map[uuid.UUID][]dtos.TicketReq)
	for _, task := range queuedTasks {
		ticketsByPool, err := u.getFormattedTickets(task.Tickets)
		if err != nil {
			log.Printf("[REQUEUE] Failed to format tickets for task %s: %v", task.ID, err)
			continue
		}
		formattedTickets[task.ID] = ticketsByPool
	}

	// Clear all queue entries so nodes are clean before re-enqueue
	for _, task := range queuedTasks {
		ticketsByPool, ok := formattedTickets[task.ID]
		if !ok {
			continue
		}
		if err := u.fallBackQueueTask(ticketsByPool); err != nil {
			log.Printf("[REQUEUE] Failed to fallback task %s: %v", task.ID, err)
		}
		if err := u.ticketRepository.DeleteQueue(task.ID); err != nil {
			log.Printf("[REQUEUE] Failed to delete queue for task %s: %v", task.ID, err)
		}
	}
	log.Printf("[REQUEUE] Cleared queue entries for nodes %v", nodeNames)

	// Re-enqueue in created_at order (FIFO)
	for _, task := range queuedTasks {
		ticketsByPool, ok := formattedTickets[task.ID]
		if !ok {
			continue
		}

		startTime, err := u.enqueueTask(ticketsByPool)
		if err != nil {
			log.Printf("[REQUEUE] Failed to re-enqueue task %s: %v", task.ID, err)
			continue
		}

		status := models.StatusQueued
		if !startTime.IsZero() && time.Until(startTime) <= time.Minute {
			status = models.StatusPending
		}

		if err := u.ticketRepository.UpdateTaskQueueInfo(task.ID, startTime, status); err != nil {
			log.Printf("[REQUEUE] Failed to update task %s queue info: %v", task.ID, err)
			continue
		}

		log.Printf("[REQUEUE] Successfully re-enqueued task %s with start time %s and status %s", task.ID, startTime, status)
	}

	return nil
}
