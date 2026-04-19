// cmd/server/main.go
package main

import (
    "context"
    "log"

    "github.com/gin-gonic/gin"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/config"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/database"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/handlers"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/middleware"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/repository"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/services"
)

func main() {
    cfg := config.Load()

    // Cloudinary
    cloudinarySvc, err := services.NewCloudinaryService(
        cfg.CloudinaryCloudName,
        cfg.CloudinaryAPIKey,
        cfg.CloudinaryAPISecret,
    )
    if err != nil {
        log.Fatalf("❌ Cloudinary init failed: %v", err)
    }

    // Repositories
    ctx := context.Background()
    firestoreClient, err := database.NewFirestoreClient(ctx, cfg.FirebaseCredentials, cfg.FirebaseCredentialsJSON)
    if err != nil {
        log.Fatalf("❌ Firebase init failed: %v", err)
    }
    defer firestoreClient.Close()
    
    productRepo := repository.NewFirebaseProductRepository(firestoreClient)
    orderRepo := repository.NewFirebaseOrderRepository(firestoreClient)
    log.Println("✅ Connected to Firebase Firestore")

    // Services
    msgService := services.NewMessageService(cfg.SellerPhone, cfg.AppURL)
    productService := services.NewProductService(productRepo)
    orderService := services.NewOrderService(orderRepo, productRepo, msgService)

    // Handlers
    productHandler := handlers.NewProductHandler(productService)
    orderHandler := handlers.NewOrderHandler(orderService)
    uploadHandler := handlers.NewUploadHandler(cloudinarySvc)

    r := gin.Default()
    r.Use(middleware.CORS(cfg.AllowedOrigins))

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Public routes
    api := r.Group("/api/v1")
    {
        api.GET("/products", productHandler.GetProducts)
        api.GET("/products/:id", productHandler.GetProduct)
        api.GET("/categories", productHandler.GetCategories)
        api.POST("/orders", orderHandler.CreateOrder)
        api.GET("/orders/track/:orderNumber", orderHandler.TrackOrder)
    }

    // Admin routes
    admin := r.Group("/api/v1/admin")
    admin.Use(middleware.AdminAuth(cfg.AdminAPIKey))
    {
        // Upload
        admin.POST("/upload", uploadHandler.UploadImage)

        // Products
        admin.GET("/products", productHandler.GetAllProductsAdmin)
        admin.POST("/products", productHandler.CreateProduct)
        admin.PUT("/products/:id", productHandler.UpdateProduct)
        admin.DELETE("/products/:id", productHandler.DeleteProduct)

        // Orders
        admin.GET("/orders", orderHandler.GetOrders)
        admin.GET("/orders/:id", orderHandler.GetOrder)
        admin.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)

        // Categories
        admin.GET("/categories", productHandler.GetCategories)
        admin.POST("/categories", productHandler.CreateCategory)
        admin.DELETE("/categories/:id", productHandler.DeleteCategory)
    }

    log.Printf("🚀 Server running on port %s", cfg.Port)
    r.Run(":" + cfg.Port)
}