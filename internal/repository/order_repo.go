package repository

import (
	"context"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetAll(ctx context.Context) ([]models.Order, error)
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)
}