package models

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID           string  `json:"id"`
	OrderID      string  `json:"order_id"`
	ProductID    string  `json:"product_id"`
	VariantID    *string `json:"variant_id,omitempty"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
	Color        string  `json:"color"`
	Size         string  `json:"size"`
	ImageURL     string  `json:"image_url"`
	Quantity     int     `json:"quantity"`
	Subtotal     float64 `json:"subtotal"`
}

type Order struct {
	ID              string      `json:"id"`
	OrderNumber     string      `json:"order_number"`
	CustomerName    string      `json:"customer_name"`
	CustomerPhone   string      `json:"customer_phone"`
	CustomerAddress string      `json:"customer_address"`
	Note            string      `json:"note"`
	TotalAmount     float64     `json:"total_amount"`
	Status          OrderStatus `json:"status"`
	PaymentMethod   string      `json:"payment_method"`
	Items           []OrderItem `json:"items"`
	MessageSent     bool        `json:"message_sent"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CreateOrderRequest struct {
	CustomerName    string             `json:"customer_name" binding:"required"`
	CustomerPhone   string             `json:"customer_phone" binding:"required"`
	CustomerAddress string             `json:"customer_address"`
	Note            string             `json:"note"`
	PaymentMethod   string             `json:"payment_method"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1"`
}

type OrderItemRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	VariantID *string `json:"variant_id"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Color     string  `json:"color"`
	Size      string  `json:"size"`
	ImageURL  string  `json:"image_url"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}