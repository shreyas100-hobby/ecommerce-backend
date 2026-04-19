package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirebaseOrderRepository struct {
	client *firestore.Client
}

func NewFirebaseOrderRepository(client *firestore.Client) *FirebaseOrderRepository {
	return &FirebaseOrderRepository{client: client}
}

const ordersCollection = "orders"

func (r *FirebaseOrderRepository) Create(ctx context.Context, order *models.Order) error {
	docRef := r.client.Collection(ordersCollection).NewDoc()
	order.ID = docRef.ID
	order.OrderNumber = fmt.Sprintf("ORD-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Assign random IDs to items
	for i := range order.Items {
		order.Items[i].ID = fmt.Sprintf("item_%d", i)
		order.Items[i].OrderID = order.ID
	}

	_, err := docRef.Set(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to create order in firestore: %w", err)
	}
	return nil
}

func (r *FirebaseOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	doc, err := r.client.Collection(ordersCollection).Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var o models.Order
	if err := doc.DataTo(&o); err != nil {
		return nil, fmt.Errorf("failed to decode order: %w", err)
	}
	o.ID = doc.Ref.ID
	return &o, nil
}

func (r *FirebaseOrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	iter := r.client.Collection(ordersCollection).OrderBy("CreatedAt", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { return nil, err }
		
		var o models.Order
		if err := doc.DataTo(&o); err == nil {
			o.ID = doc.Ref.ID
			orders = append(orders, o)
		}
	}
	return orders, nil
}

func (r *FirebaseOrderRepository) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	_, err := r.client.Collection(ordersCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "Status", Value: status},
		{Path: "UpdatedAt", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

func (r *FirebaseOrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	iter := r.client.Collection(ordersCollection).Where("OrderNumber", "==", orderNumber).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search order: %w", err)
	}

	var o models.Order
	if err := doc.DataTo(&o); err != nil {
		return nil, fmt.Errorf("failed to decode order: %w", err)
	}
	o.ID = doc.Ref.ID
	return &o, nil
}
