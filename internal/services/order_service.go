package services

import (
	"context"
	"fmt"
	"log"

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
		CustomerEmail:   req.CustomerEmail,
		CustomerAddress: req.CustomerAddress,
		GoogleMapsLink:  req.GoogleMapsLink,
		Note:            req.Note,
		PaymentMethod:   req.PaymentMethod,
		Status:          models.StatusPending,
	}

	var totalAmount float64

	// 1. Validate products and build order items
	for idx, itemReq := range req.Items {
		log.Printf("🔍 Processing item %d: ProductID=%s, VariantID=%v, Quantity=%d", 
			idx+1, itemReq.ProductID, itemReq.VariantID, itemReq.Quantity)

		product, err := s.productRepo.GetByID(ctx, itemReq.ProductID)
		if err != nil {
			log.Printf("❌ Product not found: %s", itemReq.ProductID)
			return nil, fmt.Errorf("product not found: %s", itemReq.ProductID)
		}

		log.Printf("📦 Product found: %s (Available: %v, Stock: %d, Variants: %d)", 
			product.Name, product.IsAvailable, product.StockQuantity, len(product.Variants))

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

		// Check if product has variants
		if len(product.Variants) > 0 {
			log.Printf("🎨 Product has %d variants", len(product.Variants))
			
			// If product has variants, user MUST select one
			if variantID == nil || *variantID == "" {
				return nil, fmt.Errorf("please select color and size for '%s'", product.Name)
			}

			// Find the variant
			var selectedVariant *models.ProductVariant
			for i := range product.Variants {
				log.Printf("   Variant %d: ID=%s, Color=%s, Size=%s, Stock=%d", 
					i, product.Variants[i].ID, product.Variants[i].Color, 
					product.Variants[i].Size, product.Variants[i].StockQuantity)
				
				if product.Variants[i].ID == *variantID {
					selectedVariant = &product.Variants[i]
					break
				}
			}

			if selectedVariant == nil {
				log.Printf("❌ Variant %s not found in product variants", *variantID)
				return nil, fmt.Errorf("selected variant not found for '%s'", product.Name)
			}

			log.Printf("✅ Variant found: %s-%s (Stock: %d, Requested: %d)", 
				selectedVariant.Color, selectedVariant.Size, selectedVariant.StockQuantity, itemReq.Quantity)

			// Check variant stock
			if selectedVariant.StockQuantity < itemReq.Quantity {
				return nil, fmt.Errorf(
					"only %d units available for '%s' in %s - %s",
					selectedVariant.StockQuantity, product.Name,
					selectedVariant.Color, selectedVariant.Size,
				)
			}

			color = selectedVariant.Color
			size = selectedVariant.Size
		} else {
			// Product has no variants - check main stock
			log.Printf("📦 Product has no variants, checking main stock: %d (Requested: %d)", 
				product.StockQuantity, itemReq.Quantity)

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

		log.Printf("✅ Item validated: %s x%d = ₹%.2f", product.Name, itemReq.Quantity, subtotal)
	}

	order.TotalAmount = totalAmount

	// 2. Create order FIRST (this assigns IDs and order number)
	log.Printf("💾 Creating order in database...")
	if err := s.orderRepo.Create(ctx, order); err != nil {
		log.Printf("❌ Failed to create order: %v", err)
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	log.Printf("✅ Order created: %s", order.OrderNumber)

	// 3. Then decrement stock (only after successful order creation)
	for _, item := range order.Items {
		var vID *string
		if item.VariantID != nil && *item.VariantID != "" {
			vID = item.VariantID
		}
		
		log.Printf("📉 Decrementing stock for %s (ProductID: %s, VariantID: %v, Qty: %d)", 
			item.ProductName, item.ProductID, vID, item.Quantity)

		err := s.productRepo.DecrementStock(ctx, item.ProductID, vID, item.Quantity)
		if err != nil {
			log.Printf("❌ Stock decrement failed: %v", err)
			return nil, fmt.Errorf("order created but stock update failed for '%s': %w. Contact support with order #%s", 
				item.ProductName, err, order.OrderNumber)
		}
		log.Printf("✅ Stock decremented for %s", item.ProductName)
	}

	// 4. Generate messages and return
	whatsappURL := s.msgService.GenerateWhatsAppURL(order)
	customerMessage := s.msgService.GenerateCustomerConfirmationMessage(order)
	orderMessage := s.msgService.GenerateOrderMessage(order)

	log.Printf("🎉 Order completed successfully: %s (Total: ₹%.2f)", order.OrderNumber, order.TotalAmount)

	return &CreateOrderResponse{
		Order:        order,
		WhatsAppURL:  whatsappURL,
		Message:      customerMessage,
		OrderMessage: orderMessage,
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