package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/services"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"order":         result.Order,
		"whatsapp_url":  result.WhatsAppURL,
		"message":       result.Message,
		"order_message": result.OrderMessage,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	orders, err := h.orderService.GetAllOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	c.JSON(http.StatusOK, gin.H{"data": orders, "count": len(orders)})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.orderService.UpdateStatus(
		c.Request.Context(), id, req.Status,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

func (h *OrderHandler) TrackOrder(c *gin.Context) {
	orderNumber := c.Param("orderNumber")
	order, err := h.orderService.TrackOrder(c.Request.Context(), orderNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found. Please check your order number.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"order_number":   order.OrderNumber,
			"customer_name":  order.CustomerName,
			"customer_phone": order.CustomerPhone,
			"customer_email": order.CustomerEmail,
			"status":         order.Status,
			"total_amount":   order.TotalAmount,
			"payment_method": order.PaymentMethod,
			"items":          order.Items,
			"created_at":     order.CreatedAt,
		},
	})
}