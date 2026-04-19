package services

import (
	"context"
	"fmt"

	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/repository"
)

type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	msgService  *MessageService
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	msgService *MessageService,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		msgService:  msgService,
	}
}

type CreateOrderResponse struct {
	Order        *models.Order `json:"order"`
	WhatsAppURL  string        `json:"whatsapp_url"`
	Message      string        `json:"message"`
	OrderMessage string        `json:"order_message"`
}

func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*CreateOrderResponse, error) {
	if req.PaymentMethod == "" {
		req.PaymentMethod = "cod"
	}

	order := &models.Order{
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		CustomerAddress: req.CustomerAddress,
		Note:            req.Note,
		PaymentMethod:   req.PaymentMethod,
		Status:          models.StatusPending,
	}

	var totalAmount float64

	for _, itemReq := range req.Items {
		product, err := s.productRepo.GetByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %s", itemReq.ProductID)
		}
		if !product.IsAvailable {
			return nil, fmt.Errorf("'%s' is currently unavailable", product.Name)
		}

		imageURL := product.ImageURL
		if len(product.Images) > 0 {
			imageURL = product.Images[0].URL
		}

		color := itemReq.Color
		size := itemReq.Size
		variantID := itemReq.VariantID

		if variantID != nil && *variantID != "" {
			variant, err := s.productRepo.GetVariantByID(ctx, *variantID)
			if err != nil {
				return nil, fmt.Errorf("variant not found for '%s'", product.Name)
			}
			if variant.StockQuantity < itemReq.Quantity {
				return nil, fmt.Errorf(
					"only %d units available for '%s' in %s - %s",
					variant.StockQuantity, product.Name,
					variant.Color, variant.Size,
				)
			}
			color = variant.Color
			size = variant.Size
		} else if len(product.Variants) > 0 {
			return nil, fmt.Errorf(
				"please select color and size for '%s'",
				product.Name,
			)
		} else {
			if product.StockQuantity < itemReq.Quantity {
				return nil, fmt.Errorf(
					"only %d units available for '%s'",
					product.StockQuantity, product.Name,
				)
			}
		}

		subtotal := product.Price * float64(itemReq.Quantity)
		totalAmount += subtotal

		order.Items = append(order.Items, models.OrderItem{
			ProductID:    product.ID,
			VariantID:    variantID,
			ProductName:  product.Name,
			ProductPrice: product.Price,
			Color:        color,
			Size:         size,
			ImageURL:     imageURL,
			Quantity:     itemReq.Quantity,
			Subtotal:     subtotal,
		})
	}

	order.TotalAmount = totalAmount

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return &CreateOrderResponse{
		Order:        order,
		WhatsAppURL:  s.msgService.GenerateWhatsAppURL(order),
		Message:      s.msgService.GenerateCustomerConfirmationMessage(order),
		OrderMessage: s.msgService.GenerateOrderMessage(order),
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	return s.orderRepo.GetAll(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error {
	return s.orderRepo.UpdateStatus(ctx, id, status)
}

func (s *OrderService) TrackOrder(ctx context.Context, orderNumber string) (*models.Order, error) {
	return s.orderRepo.GetByOrderNumber(ctx, orderNumber)
}