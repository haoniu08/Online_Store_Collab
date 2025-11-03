package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"CS6650_Online_Store/internal/models"
)

// DynamoDBCartRepository implements CartRepository using DynamoDB
type DynamoDBCartRepository struct {
	client    *dynamodb.DynamoDB
	tableName string
}

// NewDynamoDBCartRepository creates a new DynamoDB-based cart repository
func NewDynamoDBCartRepository(client *dynamodb.DynamoDB, tableName string) *DynamoDBCartRepository {
	return &DynamoDBCartRepository{
		client:    client,
		tableName: tableName,
	}
}

// dynamoCart represents the DynamoDB item structure
type dynamoCart struct {
	ShoppingCartID int64      `dynamodbav:"shopping_cart_id"`
	CustomerID     int64      `dynamodbav:"customer_id"`
	Status         string     `dynamodbav:"status"`
	Items          []cartItem `dynamodbav:"items"`
	CreatedAt      string     `dynamodbav:"created_at"`
	UpdatedAt      string     `dynamodbav:"updated_at"`
}

// cartItem represents an item in the shopping cart
type cartItem struct {
	ProductID int64 `dynamodbav:"product_id"`
	Quantity  int   `dynamodbav:"quantity"`
}

// CreateCart creates a new shopping cart in DynamoDB
func (r *DynamoDBCartRepository) CreateCart(ctx context.Context, customerID int64) (int64, error) {
	// Generate cart ID using timestamp + random component
	// In production, you might use a distributed ID generator
	cartID := time.Now().UnixNano() / 1000000 // Milliseconds since epoch

	now := time.Now().UTC().Format(time.RFC3339)

	// Create cart object with embedded items list
	// IMPORTANT: Cannot use MarshalMap here because empty Go slice []cartItem{} marshals to DynamoDB NULL
	// This causes GetCart to fail when unmarshaling NULL back to []cartItem
	// Solution: Manually construct item with empty List {L: []} instead of NULL
	item := map[string]*dynamodb.AttributeValue{
		"shopping_cart_id": {N: aws.String(strconv.FormatInt(cartID, 10))},
		"customer_id":      {N: aws.String(strconv.FormatInt(customerID, 10))},
		"status":           {S: aws.String("OPEN")},
		"items":            {L: []*dynamodb.AttributeValue{}}, // Empty List, NOT NULL
		"created_at":       {S: aws.String(now)},
		"updated_at":       {S: aws.String(now)},
	}

	// PutItem creates a new item in DynamoDB
	_, err := r.client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(shopping_cart_id)"),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create cart: %w", err)
	}

	return cartID, nil
}

// GetCart retrieves a shopping cart by ID from DynamoDB
func (r *DynamoDBCartRepository) GetCart(ctx context.Context, cartID int64) (*models.Cart, error) {
	// Construct the key for GetItem
	result, err := r.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"shopping_cart_id": {
				N: aws.String(strconv.FormatInt(cartID, 10)), // Number 类型存储为字符串
			},
		},
		// 设置为 true 可以强一致性读取，但延迟稍高，成本翻倍
		ConsistentRead: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Check if cart exists
	if result.Item == nil {
		return nil, ErrCartNotFound
	}

	// Unmarshal DynamoDB item to Go struct
	var cart dynamoCart
	err = dynamodbattribute.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	// Convert to API model
	return r.toDomainModel(&cart), nil
}

// AddOrUpdateItem adds or updates an item in a shopping cart
func (r *DynamoDBCartRepository) AddOrUpdateItem(ctx context.Context, cartID int64, productID int64, quantity int) error {
	// First, get the current cart to check status and get items list
	// 需要读取整个 List，修改后写回（read-modify-write pattern）
	cart, err := r.GetCart(ctx, cartID)
	if err != nil {
		return err
	}

	// Check cart status
	if cart.Status != "OPEN" {
		return errors.New("cannot modify cart with status: " + cart.Status)
	}

	// Update items list: find product and update quantity, or append new item
	found := false
	items := make([]cartItem, 0, len(cart.Items)+1)

	for _, item := range cart.Items {
		if item.ProductID == productID {
			// Update existing item's quantity (replace, not increment)
			items = append(items, cartItem{
				ProductID: productID,
				Quantity:  quantity,
			})
			found = true
		} else {
			// Keep other items unchanged
			items = append(items, cartItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
	}

	// Add new item if not found
	if !found {
		items = append(items, cartItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	// Marshal the updated items list
	itemsAttr, err := dynamodbattribute.MarshalList(items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// UpdateItem to replace the entire items list
	// SET: 更新或创建属性
	// :placeholder: 表达式属性值（防止注入，类似 SQL 的 ?）
	_, err = r.client.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"shopping_cart_id": {
				N: aws.String(strconv.FormatInt(cartID, 10)),
			},
		},
		UpdateExpression: aws.String("SET #items = :items, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":items":      {L: itemsAttr},
			":updated_at": {S: aws.String(now)},
			":open":       {S: aws.String("OPEN")},
		},
		// 原子性保证：如果状态在读取后被其他请求修改，此更新会失败
		ConditionExpression: aws.String("#status = :open"),
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("status"), // 使用 # 避免 DynamoDB 保留关键字冲突
			"#items":  aws.String("items"),  // items is also a reserved keyword
		},
	})
	if err != nil {
		// Check if condition failed
		if isConditionCheckFailed(err) {
			return errors.New("cart is not in OPEN status")
		}
		return fmt.Errorf("failed to update cart items: %w", err)
	}

	return nil
}

// toDomainModel converts DynamoDB item to domain model
func (r *DynamoDBCartRepository) toDomainModel(cart *dynamoCart) *models.Cart {
	// Convert items
	items := make([]models.CartItem, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = models.CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, cart.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, cart.UpdatedAt)

	return &models.Cart{
		ShoppingCartID: cart.ShoppingCartID,
		CustomerID:     cart.CustomerID,
		Status:         cart.Status,
		Items:          items,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}

// isConditionCheckFailed checks if error is ConditionalCheckFailedException
func isConditionCheckFailed(err error) bool {
	if err == nil {
		return false
	}
	// AWS SDK v1 error checking
	return err.Error() == "ConditionalCheckFailedException"
}
