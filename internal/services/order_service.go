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
	log.Printf("\n========== CREATE ORDER START ==========")
	log.Printf("Customer: %s (%s)", req.CustomerName, req.CustomerPhone)
	
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

	// STEP 1: Validate all items
	log.Printf("\n--- STEP 1: Validating %d items ---", len(req.Items))
	for idx, itemReq := range req.Items {
		log.Printf("\n[Item %d/%d]", idx+1, len(req.Items))
		log.Printf("  ProductID: %s", itemReq.ProductID)
		log.Printf("  VariantID: %v", itemReq.VariantID)
		log.Printf("  Quantity: %d", itemReq.Quantity)
		log.Printf("  Color: %s, Size: %s", itemReq.Color, itemReq.Size)

		// Get product
		product, err := s.productRepo.GetByID(ctx, itemReq.ProductID)
		if err != nil {
			log.Printf("  ❌ Product not found")
			return nil, fmt.Errorf("product not found: %s", itemReq.ProductID)
		}

		log.Printf("  ✅ Product: %s", product.Name)
		log.Printf("     Available: %v, Stock: %d, Variants: %d", 
			product.IsAvailable, product.StockQuantity, len(product.Variants))

		if !product.IsAvailable {
			return nil, fmt.Errorf("'%s' is currently unavailable", product.Name)
		}

		// Set defaults
		imageURL := product.ImageURL
		if len(product.Images) > 0 {
			imageURL = product.Images[0].URL
		}

		color := itemReq.Color
		size := itemReq.Size
		variantID := itemReq.VariantID

		// Handle variants
		if len(product.Variants) > 0 {
			log.Printf("     Product has variants")
			
			if variantID == nil || *variantID == "" {
				return nil, fmt.Errorf("please select color and size for '%s'", product.Name)
			}

			// Find variant
			var selectedVariant *models.ProductVariant
			for i := range product.Variants {
				log.Printf("     Checking variant: %s (%s-%s, Stock: %d)", 
					product.Variants[i].ID, product.Variants[i].Color, 
					product.Variants[i].Size, product.Variants[i].StockQuantity)
				
				if product.Variants[i].ID == *variantID {
					selectedVariant = &product.Variants[i]
					break
				}
			}

			if selectedVariant == nil {
				log.Printf("  ❌ Variant %s not found", *variantID)
				return nil, fmt.Errorf("selected variant not found for '%s'", product.Name)
			}

			log.Printf("  ✅ Variant found: %s-%s (Stock: %d)", 
				selectedVariant.Color, selectedVariant.Size, selectedVariant.StockQuantity)

			if selectedVariant.StockQuantity < itemReq.Quantity {
				return nil, fmt.Errorf("only %d units available for '%s' in %s-%s",
					selectedVariant.StockQuantity, product.Name,
					selectedVariant.Color, selectedVariant.Size)
			}

			color = selectedVariant.Color
			size = selectedVariant.Size
		} else {
			log.Printf("     No variants, checking main stock")
			
			if product.StockQuantity < itemReq.Quantity {
				return nil, fmt.Errorf("only %d units available for '%s'",
					product.StockQuantity, product.Name)
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

		log.Printf("  ✅ Validated: %s x%d = ₹%.2f", product.Name, itemReq.Quantity, subtotal)
	}

	order.TotalAmount = totalAmount
	log.Printf("\n💰 Total Amount: ₹%.2f", totalAmount)

	// STEP 2: Create order in database
	log.Printf("\n--- STEP 2: Creating order ---")
	if err := s.orderRepo.Create(ctx, order); err != nil {
		log.Printf("❌ Failed to create order: %v", err)
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	log.Printf("✅ Order created: %s", order.OrderNumber)

	// STEP 3: Decrement stock
	log.Printf("\n--- STEP 3: Decrementing stock for %d items ---", len(order.Items))
	for idx, item := range order.Items {
		log.Printf("\n[Stock Update %d/%d]", idx+1, len(order.Items))
		log.Printf("  Product: %s (ID: %s)", item.ProductName, item.ProductID)
		log.Printf("  Variant: %v", item.VariantID)
		log.Printf("  Quantity: %d", item.Quantity)

		err := s.productRepo.DecrementStock(ctx, item.ProductID, item.VariantID, item.Quantity)
		if err != nil {
			log.Printf("  ❌ STOCK DECREMENT FAILED: %v", err)
			return nil, fmt.Errorf("order #%s created but stock update failed for '%s': %w", 
				order.OrderNumber, item.ProductName, err)
		}

		log.Printf("  ✅ Stock decremented")

		// Verify the update
		updatedProduct, verifyErr := s.productRepo.GetByID(ctx, item.ProductID)
		if verifyErr == nil {
			if item.VariantID != nil && *item.VariantID != "" {
				for _, v := range updatedProduct.Variants {
					if v.ID == *item.VariantID {
						log.Printf("  🔍 VERIFIED: Variant %s now has stock=%d", v.ID, v.StockQuantity)
						break
					}
				}
			} else {
				log.Printf("  🔍 VERIFIED: Product now has stock=%d", updatedProduct.StockQuantity)
			}
		}
	}

	// STEP 4: Generate messages
	log.Printf("\n--- STEP 4: Generating messages ---")
	whatsappURL := s.msgService.GenerateWhatsAppURL(order)
	customerMessage := s.msgService.GenerateCustomerConfirmationMessage(order)
	orderMessage := s.msgService.GenerateOrderMessage(order)

	log.Printf("✅ ORDER COMPLETED SUCCESSFULLY")
	log.Printf("========== CREATE ORDER END ==========\n")

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