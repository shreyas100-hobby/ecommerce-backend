package services

import (
	"context"

	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/repository"
)

type ProductService struct {
	productRepo *repository.ProductRepository
}

func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetAll(ctx context.Context, categoryID string) ([]models.Product, error) {
	return s.productRepo.GetAll(ctx, categoryID)
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	return s.productRepo.Create(ctx, req)
}

func (s *ProductService) Update(ctx context.Context, id string, req *models.UpdateProductRequest) (*models.Product, error) {
	return s.productRepo.Update(ctx, id, req)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.productRepo.Delete(ctx, id)
}

func (s *ProductService) GetCategories(ctx context.Context) ([]models.Category, error) {
	return s.productRepo.GetAllCategories(ctx)
}
func (s *ProductService) CreateCategory(ctx context.Context, name, description string, cat *models.Category) error {
	return s.productRepo.CreateCategory(ctx, name, description, cat)
}

func (s *ProductService) DeleteCategory(ctx context.Context, id string) error {
	return s.productRepo.DeleteCategory(ctx, id)
}

func (s *ProductService) GetAllAdmin(ctx context.Context) ([]models.Product, error) {
	return s.productRepo.GetAllAdmin(ctx)
}