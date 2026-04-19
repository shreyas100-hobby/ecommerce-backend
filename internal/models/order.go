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
	ID           string  `json:"id" firestore:"id"`
	OrderID      string  `json:"order_id" firestore:"order_id"`
	ProductID    string  `json:"product_id" firestore:"product_id"`
	VariantID    *string `json:"variant_id,omitempty" firestore:"variant_id,omitempty"`
	ProductName  string  `json:"product_name" firestore:"product_name"`
	ProductPrice float64 `json:"product_price" firestore:"product_price"`
	Color        string  `json:"color" firestore:"color"`
	Size         string  `json:"size" firestore:"size"`
	ImageURL     string  `json:"image_url" firestore:"image_url"`
	Quantity     int     `json:"quantity" firestore:"quantity"`
	Subtotal     float64 `json:"subtotal" firestore:"subtotal"`
}

type Order struct {
	ID              string      `json:"id" firestore:"id"`
	OrderNumber     string      `json:"order_number" firestore:"order_number"`
	CustomerName    string      `json:"customer_name" firestore:"customer_name"`
	CustomerPhone   string      `json:"customer_phone" firestore:"customer_phone"`
	CustomerEmail   string      `json:"customer_email" firestore:"customer_email"`
	CustomerAddress string      `json:"customer_address" firestore:"customer_address"`
	GoogleMapsLink  string      `json:"google_maps_link" firestore:"google_maps_link"`
	Note            string      `json:"note" firestore:"note"`
	TotalAmount     float64     `json:"total_amount" firestore:"total_amount"`
	Status          OrderStatus `json:"status" firestore:"status"`
	PaymentMethod   string      `json:"payment_method" firestore:"payment_method"`
	Items           []OrderItem `json:"items" firestore:"items"`
	MessageSent     bool        `json:"message_sent" firestore:"message_sent"`
	CreatedAt       time.Time   `json:"created_at" firestore:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" firestore:"updated_at"`
}

type CreateOrderRequest struct {
	CustomerName    string             `json:"customer_name" binding:"required"`
	CustomerPhone   string             `json:"customer_phone" binding:"required"`
	CustomerEmail   string             `json:"customer_email"`
	CustomerAddress string             `json:"customer_address"`
	GoogleMapsLink  string             `json:"google_maps_link"`
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