package main

import (
    "context"
    "encoding/json"
    "testing"

    "CS6650_Online_Store/internal/models"
    "github.com/aws/aws-lambda-go/events"
)

func TestHandler_SimpleSNS(t *testing.T) {
    ord := models.Order{
        OrderID:    "",
        CustomerID: 123,
        Items: []models.Item{
            {ProductID: 1, Quantity: 1, Price: 9.99},
        },
    }
    b, err := json.Marshal(ord)
    if err != nil {
        t.Fatalf("marshal: %v", err)
    }

    ev := events.SNSEvent{
        Records: []events.SNSEventRecord{
            {
                SNS: events.SNSEntity{Message: string(b)},
            },
        },
    }

    if err := Handler(context.Background(), ev); err != nil {
        t.Fatalf("handler returned error: %v", err)
    }
}
