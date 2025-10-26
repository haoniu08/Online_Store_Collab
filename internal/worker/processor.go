package worker

import (
	"CS6650_Online_Store/internal/models"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// OrderProcessor processes orders from SQS queue
type OrderProcessor struct {
	sqsClient   *sqs.SQS
	queueURL    string
	workerCount int // Number of concurrent worker goroutines

	// Payment gateway bottleneck - same as synchronous handler
	// This simulates the real payment processor limitation
	paymentGateway chan struct{}

	// WaitGroup to track active workers
	wg sync.WaitGroup

	// Channel to signal shutdown
	shutdown chan struct{}
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor() (*OrderProcessor, error) {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Println("Warning: SQS_QUEUE_URL not set, order processor will not start")
		return nil, nil
	}

	// Get worker count from environment variable, default to 1
	workerCount := 1
	if workerCountStr := os.Getenv("WORKER_COUNT"); workerCountStr != "" {
		if count, err := strconv.Atoi(workerCountStr); err == nil && count > 0 {
			workerCount = count
		}
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, err
	}

	processor := &OrderProcessor{
		sqsClient:      sqs.New(sess),
		queueURL:       queueURL,
		workerCount:    workerCount,
		paymentGateway: make(chan struct{}, 1), // Buffer size 1 = only 1 payment at a time
		shutdown:       make(chan struct{}),
	}

	log.Printf("Order processor initialized - Queue: %s, Workers: %d", queueURL, workerCount)
	return processor, nil
}

// Start begins processing orders from SQS
// This is the main loop that continuously polls SQS and spawns worker goroutines
func (p *OrderProcessor) Start() {
	if p == nil {
		log.Println("Order processor not initialized, skipping...")
		return
	}

	log.Printf("Starting order processor with %d workers...", p.workerCount)

	// Main polling loop
	for {
		select {
		case <-p.shutdown:
			log.Println("Shutdown signal received, waiting for workers to finish...")
			p.wg.Wait()
			log.Println("All workers finished, processor stopped")
			return
		default:
			// Poll SQS for messages
			p.pollAndProcess()
		}
	}
}

// pollAndProcess polls SQS once and processes received messages
func (p *OrderProcessor) pollAndProcess() {
	// ReceiveMessage configuration as per assignment:
	// - WaitTimeSeconds: 20 (long polling)
	// - MaxNumberOfMessages: 10 (receive up to 10 messages)
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(p.queueURL),
		MaxNumberOfMessages: aws.Int64(10),               // Receive up to 10 messages
		WaitTimeSeconds:     aws.Int64(20),               // Long polling - wait up to 20s
		VisibilityTimeout:   aws.Int64(30),               // 30 seconds to process before message becomes visible again
		MessageAttributeNames: aws.StringSlice([]string{
			"All", // Receive all message attributes
		}),
	}

	result, err := p.sqsClient.ReceiveMessage(input)
	if err != nil {
		log.Printf("Error receiving messages from SQS: %v", err)
		time.Sleep(5 * time.Second) // Back off on error
		return
	}

	// Process each message in a separate goroutine
	for _, message := range result.Messages {
		// Increment wait group before spawning goroutine
		p.wg.Add(1)

		// Spawn goroutine to process this message
		go p.processMessage(message)
	}
}

// processMessage processes a single SQS message (one order)
func (p *OrderProcessor) processMessage(message *sqs.Message) {
	defer p.wg.Done()

	// Parse order from message body
	var order models.Order
	if err := json.Unmarshal([]byte(*message.Body), &order); err != nil {
		log.Printf("Failed to parse order from message: %v", err)
		// Don't delete message - let it become visible again for retry
		return
	}

	log.Printf("Processing order %s (customer %d) with %d items",
		order.OrderID, order.CustomerID, len(order.Items))

	// Simulate payment processing with bottleneck
	// This is the same 3-second bottleneck as synchronous processing
	startTime := time.Now()

	// Acquire payment gateway lock (blocks if busy)
	p.paymentGateway <- struct{}{}
	log.Printf("Order %s acquired payment gateway lock", order.OrderID)

	// Simulate 3-second payment verification
	time.Sleep(3 * time.Second)

	// Release lock
	<-p.paymentGateway

	processingTime := time.Since(startTime)
	log.Printf("Order %s payment completed in %v", order.OrderID, processingTime)

	// Update order status
	order.Status = models.StatusCompleted

	// Delete message from SQS (order processed successfully)
	deleteInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(p.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	}

	if _, err := p.sqsClient.DeleteMessage(deleteInput); err != nil {
		log.Printf("Failed to delete message for order %s: %v", order.OrderID, err)
		// Message will become visible again and be reprocessed
		return
	}

	log.Printf("Order %s completed and removed from queue", order.OrderID)
}

// Stop gracefully stops the processor
func (p *OrderProcessor) Stop() {
	if p != nil {
		close(p.shutdown)
	}
}

// SetWorkerCount updates the number of concurrent workers
// This is used for Phase 5 scaling experiments
func (p *OrderProcessor) SetWorkerCount(count int) {
	if count > 0 {
		p.workerCount = count
		log.Printf("Worker count updated to %d", count)
	}
}
