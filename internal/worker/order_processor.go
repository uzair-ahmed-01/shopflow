package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// OrderProcessor coordinates background order status transition workers.
type OrderProcessor struct {
	orderRepo  repository.OrderRepository
	numWorkers int
	jobChan    chan int
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewOrderProcessor creates a new OrderProcessor instance.
func NewOrderProcessor(orderRepo repository.OrderRepository, numWorkers int) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:  orderRepo,
		numWorkers: numWorkers,
		jobChan:    make(chan int, 100),
		stopChan:   make(chan struct{}),
	}
}

// Start spawns the worker pool and dispatcher loops.
func (p *OrderProcessor) Start(ctx context.Context) {
	// 1. Spawn worker pool
	for i := 1; i <= p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	// 2. Start dispatcher loop
	p.wg.Add(1)
	go p.dispatcher(ctx)

	log.Printf("Order processor background workers started successfully (workers: %d)", p.numWorkers)
}

// Stop shuts down dispatcher, closes job queue, and awaits active worker cleanups.
func (p *OrderProcessor) Stop() {
	close(p.stopChan)
	close(p.jobChan)
	p.wg.Wait()
	log.Println("Order processor background workers stopped gracefully.")
}

// Dispatcher Polling for pending orders and enqueue job to process
func (p *OrderProcessor) dispatcher(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[DISPATCHER] Context cancelled. Stopping dispatcher...")
			return

		case <-p.stopChan:
			log.Println("[DISPATCHER] Stop signal received. Stopping dispatcher...")
			return

		case <-ticker.C:
			// Fetch only eligible pending orders.
			// Repository is responsible for filtering
			// (e.g. status=PENDING and created_at older than 10 seconds).
			pendingOrders, err := p.orderRepo.ListPendingOrders(ctx)
			if err != nil {
				log.Printf("[DISPATCHER] Error fetching pending orders: %v", err)
				continue
			}

			if len(pendingOrders) == 0 {
				log.Println("[DISPATCHER] No pending orders found.")
				continue
			}

			log.Printf("[DISPATCHER] Found %d pending orders", len(pendingOrders))

			for _, order := range pendingOrders {
				select {
				case p.jobChan <- order.ID:
					log.Printf("[DISPATCHER] Enqueued Order #%d", order.ID)

				default:
					log.Printf("[DISPATCHER] Job queue full. Skipping Order #%d", order.ID)
				}
			}
		}
	}
}

func (p *OrderProcessor) worker(ctx context.Context, workerID int) {
	defer p.wg.Done()

	// Range loop terminates automatically when jobChan is closed and drained
	for orderID := range p.jobChan {
		select {
		case <-ctx.Done():
			return
		default:
			log.Printf("[WORKER %d] Processing order ID %d: PENDING -> IN_PROGRESS...", workerID, orderID)

			// Simulate processing work/delay
			time.Sleep(5 * time.Second)

			// Transition status in DB
			err := p.orderRepo.UpdateOrderStatus(ctx, orderID, models.StatusPending, models.StatusProcessing)
			if err != nil {
				log.Printf("[WORKER %d] Failed to transition order ID %d: %v", workerID, orderID, err)
				continue
			}

			log.Printf("[WORKER %d] Successfully updated status for order ID %d", workerID, orderID)
		}
	}
}
