package main

import (
	"CS6650_Online_Store/internal/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("=== Order Processor Starting ===")

	// Create order processor
	processor, err := worker.NewOrderProcessor()
	if err != nil {
		log.Fatalf("Failed to create order processor: %v", err)
	}

	if processor == nil {
		log.Println("Order processor not configured (SQS_QUEUE_URL not set)")
		log.Println("This is expected for local development without AWS resources")
		return
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start processor in a goroutine
	go processor.Start()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received...")

	// Stop processor gracefully
	processor.Stop()

	log.Println("Order processor stopped")
}
