package repository

import (
	"context"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
)

type ProductRepository interface {
	GetAll(ctx context.Context, categoryID string) ([]models.Product, error)
	GetAllAdmin(ctx context.Context) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
	Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error)
	Update(ctx context.Context, id string, req *models.UpdateProductRequest) (*models.Product, error)
	Delete(ctx context.Context, id string) error
	GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error)
	GetProductVariants(ctx context.Context, productID string) ([]models.ProductVariant, error)
	GetVariantByID(ctx context.Context, id string) (*models.ProductVariant, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, name, description string, cat *models.Category) error
	DeleteCategory(ctx context.Context, id string) error
	DecrementStock(ctx context.Context, productID string, variantID *string, quantity int) error
}