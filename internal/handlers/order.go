package handlers

import (
	"CS6650_Online_Store/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
)

type OrderHandler struct {
	// Buffered channel to simulate payment processor bottleneck
	// Only 1 payment can be processed at a time (3 seconds each)
	paymentGateway chan struct{}

	// AWS SNS client for publishing order events
	snsClient  *sns.SNS
	snsTopicArn string
}

// NewOrderHandler creates a new order handler with payment gateway simulation and AWS SNS
func NewOrderHandler() *OrderHandler {
	handler := &OrderHandler{
		// Buffer size of 1 means only 1 payment can process at a time
		paymentGateway: make(chan struct{}, 1),
	}

	// Initialize SNS client if topic ARN is provided
	snsTopicArn := os.Getenv("SNS_TOPIC_ARN")
	if snsTopicArn != "" {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION")),
		})
		if err != nil {
			log.Printf("Warning: Failed to create AWS session: %v", err)
		} else {
			handler.snsClient = sns.New(sess)
			handler.snsTopicArn = snsTopicArn
			log.Printf("SNS client initialized with topic: %s", snsTopicArn)
		}
	}

	return handler
}

// ProcessOrderSync handles POST /orders/sync
// This is the synchronous approach - customer waits for payment verification
func (h *OrderHandler) ProcessOrderSync(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Invalid request body", err.Error())
		return
	}

	// Generate order ID if not provided
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}

	// Set initial status and timestamp
	order.Status = models.StatusProcessing
	order.CreatedAt = time.Now()

	// Simulate payment processing bottleneck
	// This blocks until payment gateway is available
	h.paymentGateway <- struct{}{} // Acquire lock (blocks if channel is full)

	// Simulate 3-second payment verification
	time.Sleep(3 * time.Second)

	<-h.paymentGateway // Release lock

	// Payment successful - mark order as completed
	order.Status = models.StatusCompleted

	// Return success response
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Order processed successfully",
		"order_id": order.OrderID,
		"status":   order.Status,
	})
}

// ProcessOrderAsync handles POST /orders/async
// This is the asynchronous approach - customer gets immediate acknowledgment
// Order is published to SNS and processed by background workers
func (h *OrderHandler) ProcessOrderAsync(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Invalid request body", err.Error())
		return
	}

	// Generate order ID if not provided
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}

	// Set initial status and timestamp
	order.Status = models.StatusPending
	order.CreatedAt = time.Now()

	// Check if SNS is configured
	if h.snsClient == nil {
		respondWithError(w, http.StatusServiceUnavailable, "SNS_NOT_CONFIGURED",
			"Async processing not available", "SNS client not initialized")
		return
	}

	// Publish order to SNS topic
	orderJSON, err := json.Marshal(order)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to serialize order", err.Error())
		return
	}

	input := &sns.PublishInput{
		Message:  aws.String(string(orderJSON)),
		TopicArn: aws.String(h.snsTopicArn),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"order_id": {
				DataType:    aws.String("String"),
				StringValue: aws.String(order.OrderID),
			},
		},
	}

	result, err := h.snsClient.Publish(input)
	if err != nil {
		log.Printf("Failed to publish to SNS: %v", err)
		respondWithError(w, http.StatusInternalServerError, "PUBLISH_FAILED",
			"Failed to queue order for processing", err.Error())
		return
	}

	log.Printf("Order %s published to SNS. MessageID: %s", order.OrderID, *result.MessageId)

	// Return 202 Accepted - order is queued for processing
	respondWithJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":    "Order accepted for processing",
		"order_id":   order.OrderID,
		"status":     order.Status,
		"message_id": *result.MessageId,
	})
}
