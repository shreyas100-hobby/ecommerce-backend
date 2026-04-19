package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order.OrderNumber = fmt.Sprintf("ORD-%s-%04d",
		time.Now().Format("20060102"),
		time.Now().UnixNano()%10000,
	)

	err = tx.QueryRow(ctx, `
		INSERT INTO orders
			(order_number, customer_name, customer_phone,
			 customer_address, note, total_amount, payment_method, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`,
		order.OrderNumber,
		order.CustomerName,
		order.CustomerPhone,
		order.CustomerAddress,
		order.Note,
		order.TotalAmount,
		order.PaymentMethod,
		order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for i := range order.Items {
		item := &order.Items[i]
		item.OrderID = order.ID

		err = tx.QueryRow(ctx, `
			INSERT INTO order_items
				(order_id, product_id, variant_id, product_name,
				 product_price, color, size, image_url, quantity, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`,
			item.OrderID,
			item.ProductID,
			item.VariantID,
			item.ProductName,
			item.ProductPrice,
			item.Color,
			item.Size,
			item.ImageURL,
			item.Quantity,
			item.Subtotal,
		).Scan(&item.ID)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order

	orderQuery := `
		SELECT id, order_number, customer_name, customer_phone, 
		       customer_address, note, total_amount, status, 
		       payment_method, message_sent, created_at, updated_at
		FROM orders WHERE id = $1
	`
	err := r.db.QueryRow(ctx, orderQuery, id).Scan(
		&order.ID, &order.OrderNumber,
		&order.CustomerName, &order.CustomerPhone,
		&order.CustomerAddress, &order.Note,
		&order.TotalAmount, &order.Status,
		&order.PaymentMethod, &order.MessageSent,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Fetch items
	itemsQuery := `
		SELECT id, order_id, product_id, product_name, product_price, quantity, subtotal
		FROM order_items WHERE order_id = $1
	`
	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID,
			&item.ProductName, &item.ProductPrice,
			&item.Quantity, &item.Subtotal,
		)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	query := `
		SELECT id, order_number, customer_name, customer_phone,
		       customer_address, note, total_amount, status,
		       payment_method, message_sent, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(
			&o.ID, &o.OrderNumber,
			&o.CustomerName, &o.CustomerPhone,
			&o.CustomerAddress, &o.Note,
			&o.TotalAmount, &o.Status,
			&o.PaymentMethod, &o.MessageSent,
			&o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	result, err := r.db.Exec(ctx,
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.QueryRow(ctx, `
		SELECT id, order_number, customer_name, customer_phone,
		       customer_address, note, total_amount, status,
		       payment_method, message_sent, created_at, updated_at
		FROM orders WHERE order_number = $1
	`, orderNumber).Scan(
		&order.ID, &order.OrderNumber,
		&order.CustomerName, &order.CustomerPhone,
		&order.CustomerAddress, &order.Note,
		&order.TotalAmount, &order.Status,
		&order.PaymentMethod, &order.MessageSent,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	order.Items, err = r.getOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
func (r *OrderRepository) getOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, order_id, product_id,
			variant_id,
			product_name, product_price,
			COALESCE(color, '') as color,
			COALESCE(size, '') as size,
			COALESCE(image_url, '') as image_url,
			quantity, subtotal
		FROM order_items
		WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID,
			&item.VariantID,
			&item.ProductName, &item.ProductPrice,
			&item.Color, &item.Size, &item.ImageURL,
			&item.Quantity, &item.Subtotal,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}