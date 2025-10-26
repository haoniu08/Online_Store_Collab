package main

import (
    "context"
    "encoding/json"
    "log"

    "CS6650_Online_Store/internal/models"
    "CS6650_Online_Store/internal/worker"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/google/uuid"
)

// Handler is the Lambda entrypoint for SNS events.
// It unmarshals the SNS message into an Order and delegates to the worker's
// processOrder logic (using a local payment gateway channel to simulate
// the payment bottleneck).
func Handler(ctx context.Context, e events.SNSEvent) error {
    // Create a short-lived processor that only provides the payment gateway
    // token channel. We don't need SQS client or queueURL in Lambda.
    p := worker.NewLocalProcessor()

    for _, record := range e.Records {
        msg := record.SNS.Message

        // SNS may wrap messages; attempt to unmarshal directly into Order
        var ord models.Order
        if err := json.Unmarshal([]byte(msg), &ord); err != nil {
            log.Printf("lambda: failed to unmarshal sns message: %v", err)
            // Return error to cause Lambda/SNS retry semantics
            return err
        }

        if ord.OrderID == "" {
            ord.OrderID = uuid.New().String()
        }

        if err := p.Process(ctx, &ord); err != nil {
            log.Printf("lambda: processing order %s failed: %v", ord.OrderID, err)
            // Return error so Lambda marks invocation as failed (and SNS/Lambda retries)
            return err
        }

        log.Printf("lambda: order %s processed", ord.OrderID)
    }

    return nil
}

func main() {
    lambda.Start(Handler)
}
