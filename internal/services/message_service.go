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
// internal/services/message_service.go
func (s *MessageService) GenerateOrderMessage(order *models.Order) string {
    var sb strings.Builder

    sb.WriteString("🛒 *NEW ORDER RECEIVED*\n")
    sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n\n")
    sb.WriteString(fmt.Sprintf("📋 *Order:* %s\n", order.OrderNumber))
    sb.WriteString(fmt.Sprintf("📅 *Date:* %s\n\n",
        order.CreatedAt.Format("02 Jan 2006, 3:04 PM")))

    sb.WriteString("👤 *Customer Details*\n")
    sb.WriteString(fmt.Sprintf("Name: %s\n", order.CustomerName))
    sb.WriteString(fmt.Sprintf("Phone: %s\n", order.CustomerPhone))
    if order.CustomerAddress != "" {
        sb.WriteString(fmt.Sprintf("Address: %s\n", order.CustomerAddress))
    }
    sb.WriteString("\n")

    sb.WriteString("🛍️ *Items Ordered*\n")
    sb.WriteString("─────────────────────\n")

    for _, item := range order.Items {
        sb.WriteString(fmt.Sprintf("👗 *%s*\n", item.ProductName))

        // Variant details
        if item.Color != "" {
            sb.WriteString(fmt.Sprintf("   Color: %s\n", item.Color))
        }
        if item.Size != "" {
            sb.WriteString(fmt.Sprintf("   Size: %s\n", item.Size))
        }

        sb.WriteString(fmt.Sprintf("   Qty: %d × ₹%.2f = ₹%.2f\n",
            item.Quantity, item.ProductPrice, item.Subtotal))

        // Image link
        if item.ImageURL != "" {
            sb.WriteString(fmt.Sprintf("   🖼 Photo: %s\n", item.ImageURL))
        }
        sb.WriteString("\n")
    }

    sb.WriteString("─────────────────────\n")
    sb.WriteString(fmt.Sprintf("💰 *Total: ₹%.2f*\n\n", order.TotalAmount))
    sb.WriteString(fmt.Sprintf("💳 *Payment:* %s\n",
        strings.ToUpper(order.PaymentMethod)))

    if order.Note != "" {
        sb.WriteString(fmt.Sprintf("\n📝 *Note:* %s\n", order.Note))
    }

    sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━")
    return sb.String()
}

func (s *MessageService) GenerateWhatsAppURL(order *models.Order) string {
	message := s.GenerateOrderMessage(order)
	encoded := url.QueryEscape(message)
	return fmt.Sprintf("https://wa.me/%s?text=%s", s.sellerPhone, encoded)
}

func (s *MessageService) GenerateCustomerConfirmationMessage(order *models.Order) string {
	return fmt.Sprintf(
		"Hi %s! 👋\n\nYour order *%s* has been placed! 🎉\n\nClick the button below to confirm your order on WhatsApp.",
		order.CustomerName,
		order.OrderNumber,
	)
}