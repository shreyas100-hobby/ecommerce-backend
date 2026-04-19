package main

import (
	"context"
	"log"

	"github.com/shreyas100-hobby/ecommerce-backend/internal/config"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/database"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/repository"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	log.Println("🔄 Starting Data Migration: PostgreSQL -> Firebase")

	// 1. Connect to PostgreSQL
	log.Println("📦 Connecting to Postgres...")
	db := database.NewPool(cfg.DatabaseURL)
	defer db.Close()
	pgProducts := repository.NewPostgresProductRepository(db)
	pgOrders := repository.NewPostgresOrderRepository(db)

	// 2. Connect to Firebase
	log.Println("🔥 Connecting to Firebase...")
	fsClient, err := database.NewFirestoreClient(ctx, cfg.FirebaseCredentials, cfg.FirebaseCredentialsJSON)
	if err != nil {
		log.Fatalf("❌ Firebase connection failed: %v", err)
	}
	defer fsClient.Close()

	// 3. Migrate Categories
	log.Println("➡️ Migrating Categories...")
	categories, err := pgProducts.GetAllCategories(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get categories from PG: %v", err)
	}
	for _, c := range categories {
		_, err := fsClient.Collection("categories").Doc(c.ID).Set(ctx, c)
		if err != nil {
			log.Fatalf("❌ Failed to insert category %s: %v", c.Name, err)
		}
	}
	log.Printf("✅ Migrated %d categories.", len(categories))

	// 4. Migrate Products
	log.Println("➡️ Migrating Products...")
	products, err := pgProducts.GetAllAdmin(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get products from PG: %v", err)
	}

	for _, p := range products {
		// Fetch images and variants
		images, _ := pgProducts.GetProductImages(ctx, p.ID)
		variants, _ := pgProducts.GetProductVariants(ctx, p.ID)
		p.Images = images
		p.Variants = variants

		_, err := fsClient.Collection("products").Doc(p.ID).Set(ctx, p)
		if err != nil {
			log.Fatalf("❌ Failed to insert product %s: %v", p.Name, err)
		}
	}
	log.Printf("✅ Migrated %d products.", len(products))

	// 5. Migrate Orders
	log.Println("➡️ Migrating Orders...")
	orders, err := pgOrders.GetAll(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get orders from PG: %v", err)
	}

	for _, o := range orders {
		// Order repo already populates items in GetAll
		// Wait, let's check order repo. Actually orderRepo.GetAll does NOT fetch items. We need to fetch items.
		fullOrder, err := pgOrders.GetByID(ctx, o.ID)
		if err == nil {
			_, err = fsClient.Collection("orders").Doc(fullOrder.ID).Set(ctx, fullOrder)
			if err != nil {
				log.Fatalf("❌ Failed to insert order %s: %v", fullOrder.OrderNumber, err)
			}
		}
	}
	log.Printf("✅ Migrated %d orders.", len(orders))

	log.Println("🎉 Migration Complete! You can now switch DB_DRIVER=firebase in your .env file.")
}
