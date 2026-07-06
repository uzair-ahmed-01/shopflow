package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// OrderProcessor coordinates background order status transition workers.
type OrderProcessor struct {
	orderRepo      repository.OrderRepository
	numWorkers     int
	jobChan        chan int
	stopChan       chan struct{}
	dispatcherDone chan struct{}
	wg             sync.WaitGroup
}

// NewOrderProcessor creates a new OrderProcessor instance.
func NewOrderProcessor(orderRepo repository.OrderRepository, numWorkers int) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:      orderRepo,
		numWorkers:     numWorkers,
		jobChan:        make(chan int, 100),
		stopChan:       make(chan struct{}),
		dispatcherDone: make(chan struct{}),
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
	go p.dispatcher(ctx)

	log.Info().Int("workers", p.numWorkers).Msg("Order processor background workers started successfully")
}

// Stop shuts down dispatcher, closes job queue, and awaits active worker cleanups.
func (p *OrderProcessor) Stop() {
	close(p.stopChan)
	<-p.dispatcherDone
	close(p.jobChan)
	p.wg.Wait()
	log.Info().Msg("Order processor background workers stopped gracefully")
}

// Dispatcher Polling for pending orders and enqueue job to process
func (p *OrderProcessor) dispatcher(ctx context.Context) {
	defer close(p.dispatcherDone)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("[DISPATCHER] Context cancelled. Stopping dispatcher...")
			return

		case <-p.stopChan:
			log.Info().Msg("[DISPATCHER] Stop signal received. Stopping dispatcher...")
			return

		case <-ticker.C:
			// Fetch only eligible pending orders.
			// Repository is responsible for filtering
			// (e.g. status=PENDING and created_at older than 10 seconds).
			pendingOrders, err := p.orderRepo.ListPendingOrders(ctx)
			if err != nil {
				log.Error().Err(err).Msg("[DISPATCHER] Error fetching pending orders")
				continue
			}

			if len(pendingOrders) == 0 {
				log.Debug().Msg("[DISPATCHER] No pending orders found.")
				continue
			}

			log.Info().Int("count", len(pendingOrders)).Msg("[DISPATCHER] Found pending orders")

			for _, order := range pendingOrders {
				select {
				case p.jobChan <- order.ID:
					log.Info().Int("order_id", order.ID).Msg("[DISPATCHER] Enqueued Order")

				default:
					log.Warn().Int("order_id", order.ID).Msg("[DISPATCHER] Job queue full. Skipping Order")
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
			log.Info().Int("worker_id", workerID).Int("order_id", orderID).Msg("[WORKER] Processing order ID: PENDING -> IN_PROGRESS...")

			// Simulate processing work/delay
			time.Sleep(5 * time.Second)

			// Transition status in DB
			err := p.orderRepo.UpdateOrderStatus(ctx, orderID, models.StatusPending, models.StatusProcessing)
			if err != nil {
				log.Error().Err(err).Int("worker_id", workerID).Int("order_id", orderID).Msg("[WORKER] Failed to transition order ID")
				continue
			}

			log.Info().Int("worker_id", workerID).Int("order_id", orderID).Msg("[WORKER] Successfully updated status for order ID")
		}
	}
}
