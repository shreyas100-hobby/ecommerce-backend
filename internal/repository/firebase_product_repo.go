package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirebaseProductRepository struct {
	client *firestore.Client
}

func NewFirebaseProductRepository(client *firestore.Client) *FirebaseProductRepository {
	return &FirebaseProductRepository{client: client}
}

const productsCollection = "products"
const categoriesCollection = "categories"

func (r *FirebaseProductRepository) GetAll(ctx context.Context, categoryID string) ([]models.Product, error) {
	var products []models.Product
	iter := r.client.Collection(productsCollection).OrderBy("created_at", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate products: %w", err)
		}
		var p models.Product
		if err := doc.DataTo(&p); err != nil {
			return nil, fmt.Errorf("failed to decode product: %w", err)
		}
		
		// Filter in memory to avoid needing Firestore Composite Indexes!
		if !p.IsAvailable {
			continue
		}
		if categoryID != "" && (p.CategoryID == nil || *p.CategoryID != categoryID) {
			continue
		}

		p.ID = doc.Ref.ID
		products = append(products, p)
	}
	return products, nil
}

func (r *FirebaseProductRepository) GetAllAdmin(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	iter := r.client.Collection(productsCollection).OrderBy("created_at", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate products: %w", err)
		}
		var p models.Product
		if err := doc.DataTo(&p); err != nil {
			return nil, fmt.Errorf("failed to decode product: %w", err)
		}
		p.ID = doc.Ref.ID
		products = append(products, p)
	}
	return products, nil
}

func (r *FirebaseProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	doc, err := r.client.Collection(productsCollection).Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	var p models.Product
	if err := doc.DataTo(&p); err != nil {
		return nil, fmt.Errorf("failed to decode product: %w", err)
	}
	p.ID = doc.Ref.ID
	return &p, nil
}

func (r *FirebaseProductRepository) Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	docRef := r.client.Collection(productsCollection).NewDoc()

	p := models.Product{
		ID:            docRef.ID,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		CategoryID:    req.CategoryID,
		ImageURL:      req.ImageURL,
		StockQuantity: req.StockQuantity,
		IsAvailable:   req.IsAvailable,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Map images
	for i, imgReq := range req.Images {
		p.Images = append(p.Images, models.ProductImage{
			ID:        fmt.Sprintf("img_%d", i),
			ProductID: p.ID,
			URL:       imgReq.URL,
			PublicID:  imgReq.PublicID,
			SortOrder: imgReq.SortOrder,
			CreatedAt: time.Now(),
		})
	}

	// Map variants
	for i, varReq := range req.Variants {
		p.Variants = append(p.Variants, models.ProductVariant{
			ID:            fmt.Sprintf("var_%d", i),
			ProductID:     p.ID,
			Color:         varReq.Color,
			Size:          varReq.Size,
			StockQuantity: varReq.StockQuantity,
			CreatedAt:     time.Now(),
		})
	}

	_, err := docRef.Set(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create product in firestore: %w", err)
	}

	return &p, nil
}

func (r *FirebaseProductRepository) Update(ctx context.Context, id string, req *models.UpdateProductRequest) (*models.Product, error) {
	docRef := r.client.Collection(productsCollection).Doc(id)
	
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			return err
		}
		
		var p models.Product
		if err := doc.DataTo(&p); err != nil {
			return err
		}

		if req.Name != nil { p.Name = *req.Name }
		if req.Description != nil { p.Description = *req.Description }
		if req.Price != nil { p.Price = *req.Price }
		if req.OriginalPrice != nil { p.OriginalPrice = req.OriginalPrice }
		if req.CategoryID != nil { p.CategoryID = req.CategoryID }
		if req.ImageURL != nil { p.ImageURL = *req.ImageURL }
		if req.StockQuantity != nil { p.StockQuantity = *req.StockQuantity }
		if req.IsAvailable != nil { p.IsAvailable = *req.IsAvailable }
		p.UpdatedAt = time.Now()

		if req.Images != nil {
			p.Images = []models.ProductImage{}
			for i, imgReq := range req.Images {
				p.Images = append(p.Images, models.ProductImage{
					ID:        fmt.Sprintf("img_%d", i),
					ProductID: id,
					URL:       imgReq.URL,
					PublicID:  imgReq.PublicID,
					SortOrder: imgReq.SortOrder,
					CreatedAt: time.Now(),
				})
			}
		}

		if req.Variants != nil {
			p.Variants = []models.ProductVariant{}
			for i, varReq := range req.Variants {
				p.Variants = append(p.Variants, models.ProductVariant{
					ID:            fmt.Sprintf("var_%d", i),
					ProductID:     id,
					Color:         varReq.Color,
					Size:          varReq.Size,
					StockQuantity: varReq.StockQuantity,
					CreatedAt:     time.Now(),
				})
			}
		}

		return tx.Set(docRef, p)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *FirebaseProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(productsCollection).Doc(id).Delete(ctx)
	return err
}

func (r *FirebaseProductRepository) GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error) {
	p, err := r.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	return p.Images, nil
}

func (r *FirebaseProductRepository) GetProductVariants(ctx context.Context, productID string) ([]models.ProductVariant, error) {
	p, err := r.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	return p.Variants, nil
}

func (r *FirebaseProductRepository) GetVariantByID(ctx context.Context, id string) (*models.ProductVariant, error) {
	// Need to search all products to find a variant... not ideal, but this is why NoSQL models differ.
	// For a quick migration, we'll implement a basic search.
	iter := r.client.Collection(productsCollection).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { continue }
		var p models.Product
		if err := doc.DataTo(&p); err == nil {
			for _, v := range p.Variants {
				if v.ID == id {
					return &v, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("variant not found")
}

func (r *FirebaseProductRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	iter := r.client.Collection(categoriesCollection).OrderBy("name", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { return nil, err }
		
		var c models.Category
		if err := doc.DataTo(&c); err == nil {
			c.ID = doc.Ref.ID
			categories = append(categories, c)
		}
	}
	return categories, nil
}

func (r *FirebaseProductRepository) CreateCategory(ctx context.Context, name, description string, cat *models.Category) error {
	docRef := r.client.Collection(categoriesCollection).NewDoc()
	cat.ID = docRef.ID
	cat.Name = name
	cat.Description = description
	cat.CreatedAt = time.Now()

	_, err := docRef.Set(ctx, cat)
	return err
}

func (r *FirebaseProductRepository) DeleteCategory(ctx context.Context, id string) error {
	_, err := r.client.Collection(categoriesCollection).Doc(id).Delete(ctx)
	return err
}

func (r *FirebaseProductRepository) DecrementStock(ctx context.Context, productID string, variantID *string, quantity int) error {
	docRef := r.client.Collection(productsCollection).Doc(productID)

	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			return fmt.Errorf("failed to get product for stock update: %w", err)
		}

		var p models.Product
		if err := doc.DataTo(&p); err != nil {
			return fmt.Errorf("failed to decode product for stock update: %w", err)
		}

		if variantID != nil && *variantID != "" {
			// Find and update variant stock
			found := false
			for i := range p.Variants {
				if p.Variants[i].ID == *variantID {
					if p.Variants[i].StockQuantity < quantity {
						return fmt.Errorf("insufficient stock for variant")
					}
					p.Variants[i].StockQuantity -= quantity
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("variant not found")
			}

			// Check if ALL variants are now out of stock
			allOutOfStock := true
			for _, v := range p.Variants {
				if v.StockQuantity > 0 {
					allOutOfStock = false
					break
				}
			}
			if allOutOfStock {
				p.IsAvailable = false
			}
		} else {
			// Update main product stock
			if p.StockQuantity < quantity {
				return fmt.Errorf("insufficient stock for product")
			}
			p.StockQuantity -= quantity
			
			// If stock is now 0, mark as unavailable for better UX
			if p.StockQuantity == 0 {
				p.IsAvailable = false
			}
		}

		// Also update IsAvailable if stock reaches 0?
		// Actually, let's keep IsAvailable as a manual toggle or logic based, 
		// but typically if all stock is 0, it should probably be shown as out of stock.
		// For now, just decrement.

		return tx.Set(docRef, p)
	})
}
