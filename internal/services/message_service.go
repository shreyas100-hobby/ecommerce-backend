package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
)

type MessageService struct {
	sellerPhone string
	appURL      string
}

func NewMessageService(sellerPhone, appURL string) *MessageService {
	return &MessageService{
		sellerPhone: sellerPhone,
		appURL:      appURL,
	}
}

func (s *MessageService) GenerateOrderMessage(order *models.Order) string {
	var builder strings.Builder
	
	builder.WriteString("🛍️ *New Order Received*\n\n")
	builder.WriteString(fmt.Sprintf("📋 Order #: %s\n", order.OrderNumber))
	builder.WriteString(fmt.Sprintf("👤 Customer: %s\n", order.CustomerName))
	builder.WriteString(fmt.Sprintf("📞 Phone: %s\n", order.CustomerPhone))
	
	if order.CustomerEmail != "" {
		builder.WriteString(fmt.Sprintf("📧 Email: %s\n", order.CustomerEmail))
	}
	
	if order.CustomerAddress != "" {
		builder.WriteString(fmt.Sprintf("📍 Address: %s\n", order.CustomerAddress))
	}
	
	if order.GoogleMapsLink != "" {
		builder.WriteString(fmt.Sprintf("🗺️ Location: %s\n", order.GoogleMapsLink))
	}
	
	builder.WriteString("\n*Items:*\n")
	for i, item := range order.Items {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.ProductName))
		if item.Color != "" {
			builder.WriteString(fmt.Sprintf("   Color: %s\n", item.Color))
		}
		if item.Size != "" {
			builder.WriteString(fmt.Sprintf("   Size: %s\n", item.Size))
		}
		builder.WriteString(fmt.Sprintf("   Qty: %d × ₹%.2f = ₹%.2f\n", 
			item.Quantity, item.ProductPrice, item.Subtotal))
	}
	
	builder.WriteString(fmt.Sprintf("\n💰 *Total Amount: ₹%.2f*\n", order.TotalAmount))
	
	if order.PaymentMethod != "" {
		builder.WriteString(fmt.Sprintf("💳 Payment: %s\n", order.PaymentMethod))
	}
	
	if order.Note != "" {
		builder.WriteString(fmt.Sprintf("\n📝 Note: %s\n", order.Note))
	}
	
	return builder.String()
}

func (s *MessageService) GenerateWhatsAppURL(order *models.Order) string {
	message := s.GenerateOrderMessage(order)
	trackingURL := fmt.Sprintf("%s/track/%s", s.appURL, order.OrderNumber)
	message += fmt.Sprintf("\n🔍 Track Order: %s", trackingURL)
	
	encodedMessage := url.QueryEscape(message)
	return fmt.Sprintf("https://wa.me/%s?text=%s", s.sellerPhone, encodedMessage)
}

func (s *MessageService) GenerateCustomerConfirmationMessage(order *models.Order) string {
	return fmt.Sprintf(
		"✅ Order confirmed! Your order #%s has been placed successfully. Total: ₹%.2f. We'll contact you shortly.",
		order.OrderNumber,
		order.TotalAmount,
	)
}