package models

import "time"

type Category struct {
	ID          string    `json:"id" firestore:"id"`
	Name        string    `json:"name" firestore:"name"`
	Description string    `json:"description" firestore:"description"`
	CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
}

type ProductImage struct {
	ID        string    `json:"id" firestore:"id"`
	ProductID string    `json:"product_id" firestore:"product_id"`
	URL       string    `json:"url" firestore:"url"`
	PublicID  string    `json:"public_id" firestore:"public_id"`
	SortOrder int       `json:"sort_order" firestore:"sort_order"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}

type ProductVariant struct {
	ID            string    `json:"id" firestore:"id"`
	ProductID     string    `json:"product_id" firestore:"product_id"`
	Color         string    `json:"color" firestore:"color"`
	Size          string    `json:"size" firestore:"size"`
	StockQuantity int       `json:"stock_quantity" firestore:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at" firestore:"created_at"`
}

type Product struct {
	ID            string           `json:"id" firestore:"id"`
	Name          string           `json:"name" firestore:"name"`
	Description   string           `json:"description" firestore:"description"`
	Price         float64          `json:"price" firestore:"price"`
	OriginalPrice *float64         `json:"original_price,omitempty" firestore:"original_price,omitempty"`
	CategoryID    *string          `json:"category_id,omitempty" firestore:"category_id,omitempty"`
	CategoryName  string           `json:"category_name,omitempty" firestore:"category_name,omitempty"`
	ImageURL      string           `json:"image_url" firestore:"image_url"`
	Images        []ProductImage   `json:"images" firestore:"images"`
	Variants      []ProductVariant `json:"variants" firestore:"variants"`
	StockQuantity int              `json:"stock_quantity" firestore:"stock_quantity"`
	IsAvailable   bool             `json:"is_available" firestore:"is_available"`
	CreatedAt     time.Time        `json:"created_at" firestore:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" firestore:"updated_at"`
}

type CreateProductRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description"`
	Price         float64                `json:"price" binding:"required,gt=0"`
	OriginalPrice *float64               `json:"original_price"`
	CategoryID    *string                `json:"category_id"`
	ImageURL      string                 `json:"image_url"`
	StockQuantity int                    `json:"stock_quantity"`
	IsAvailable   bool                   `json:"is_available"`
	Images        []CreateImageRequest   `json:"images"`
	Variants      []CreateVariantRequest `json:"variants"`
}

type CreateImageRequest struct {
	URL       string `json:"url"`
	PublicID  string `json:"public_id"`
	SortOrder int    `json:"sort_order"`
}

type CreateVariantRequest struct {
	Color         string `json:"color" binding:"required"`
	Size          string `json:"size" binding:"required"`
	StockQuantity int    `json:"stock_quantity"`
}

type UpdateProductRequest struct {
	Name          *string                `json:"name"`
	Description   *string                `json:"description"`
	Price         *float64               `json:"price"`
	OriginalPrice *float64               `json:"original_price"`
	CategoryID    *string                `json:"category_id"`
	ImageURL      *string                `json:"image_url"`
	StockQuantity *int                   `json:"stock_quantity"`
	IsAvailable   *bool                  `json:"is_available"`
	Images        []CreateImageRequest   `json:"images"`
	Variants      []CreateVariantRequest `json:"variants"`
}