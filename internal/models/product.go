package models

import "time"

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	URL       string    `json:"url"`
	PublicID  string    `json:"public_id"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductVariant struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	Color         string    `json:"color"`
	Size          string    `json:"size"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
}

type Product struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Price         float64          `json:"price"`
	OriginalPrice *float64         `json:"original_price,omitempty"`
	CategoryID    *string          `json:"category_id,omitempty"`
	CategoryName  string           `json:"category_name,omitempty"`
	ImageURL      string           `json:"image_url"`
	Images        []ProductImage   `json:"images"`
	Variants      []ProductVariant `json:"variants"`
	StockQuantity int              `json:"stock_quantity"`
	IsAvailable   bool             `json:"is_available"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
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