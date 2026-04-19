package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shreyas100-hobby/ecommerce-backend/internal/models"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

// ─── Products ────────────────────────────────────────────────

func (r *ProductRepository) GetAll(ctx context.Context, categoryID string) ([]models.Product, error) {
	query := `
		SELECT
			p.id, p.name, p.description, p.price, p.original_price,
			p.category_id, COALESCE(c.name, '') as category_name,
			COALESCE(p.image_url, ''),
			p.stock_quantity, p.is_available, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.is_available = true
	`
	args := []interface{}{}

	if categoryID != "" {
		query += " AND p.category_id = $1"
		args = append(args, categoryID)
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice,
			&p.CategoryID, &p.CategoryName,
			&p.ImageURL,
			&p.StockQuantity, &p.IsAvailable,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetAllAdmin(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT
			p.id, p.name, p.description, p.price, p.original_price,
			p.category_id, COALESCE(c.name, '') as category_name,
			COALESCE(p.image_url, ''),
			p.stock_quantity, p.is_available, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice,
			&p.CategoryID, &p.CategoryName,
			&p.ImageURL,
			&p.StockQuantity, &p.IsAvailable,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT
			p.id, p.name, p.description, p.price, p.original_price,
			p.category_id, COALESCE(c.name, '') as category_name,
			COALESCE(p.image_url, ''),
			p.stock_quantity, p.is_available, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`
	var p models.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice,
		&p.CategoryID, &p.CategoryName,
		&p.ImageURL,
		&p.StockQuantity, &p.IsAvailable,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	images, err := r.GetProductImages(ctx, p.ID)
	if err == nil {
		p.Images = images
	}

	variants, err := r.GetProductVariants(ctx, p.ID)
	if err == nil {
		p.Variants = variants
	}

	return &p, nil
}

func (r *ProductRepository) Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var p models.Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products
			(name, description, price, original_price, category_id,
			 image_url, stock_quantity, is_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, description, price, original_price,
		          category_id, COALESCE(image_url, ''),
		          stock_quantity, is_available, created_at, updated_at
	`,
		req.Name, req.Description, req.Price, req.OriginalPrice,
		req.CategoryID, req.ImageURL, req.StockQuantity, req.IsAvailable,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice,
		&p.CategoryID, &p.ImageURL,
		&p.StockQuantity, &p.IsAvailable,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	for i, img := range req.Images {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_images (product_id, url, public_id, sort_order)
			VALUES ($1, $2, $3, $4)
		`, p.ID, img.URL, img.PublicID, i)
		if err != nil {
			return nil, fmt.Errorf("failed to insert image: %w", err)
		}
	}

	for _, v := range req.Variants {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_variants (product_id, color, size, stock_quantity)
			VALUES ($1, $2, $3, $4)
		`, p.ID, v.Color, v.Size, v.StockQuantity)
		if err != nil {
			return nil, fmt.Errorf("failed to insert variant: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	p.Images, _ = r.GetProductImages(ctx, p.ID)
	p.Variants, _ = r.GetProductVariants(ctx, p.ID)

	return &p, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, req *models.UpdateProductRequest) (*models.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var p models.Product
	err = tx.QueryRow(ctx, `
		UPDATE products SET
			name           = COALESCE($1, name),
			description    = COALESCE($2, description),
			price          = COALESCE($3, price),
			original_price = COALESCE($4, original_price),
			category_id    = COALESCE($5, category_id),
			image_url      = COALESCE($6, image_url),
			stock_quantity = COALESCE($7, stock_quantity),
			is_available   = COALESCE($8, is_available),
			updated_at     = NOW()
		WHERE id = $9
		RETURNING id, name, description, price, original_price,
		          category_id, COALESCE(image_url, ''),
		          stock_quantity, is_available, created_at, updated_at
	`,
		req.Name, req.Description, req.Price, req.OriginalPrice,
		req.CategoryID, req.ImageURL, req.StockQuantity, req.IsAvailable,
		id,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice,
		&p.CategoryID, &p.ImageURL,
		&p.StockQuantity, &p.IsAvailable,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	if len(req.Images) > 0 {
		_, err = tx.Exec(ctx,
			"DELETE FROM product_images WHERE product_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete old images: %w", err)
		}
		for i, img := range req.Images {
			_, err = tx.Exec(ctx, `
				INSERT INTO product_images (product_id, url, public_id, sort_order)
				VALUES ($1, $2, $3, $4)
			`, p.ID, img.URL, img.PublicID, i)
			if err != nil {
				return nil, fmt.Errorf("failed to insert image: %w", err)
			}
		}
	}

	if len(req.Variants) > 0 {
		_, err = tx.Exec(ctx,
			"DELETE FROM product_variants WHERE product_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete old variants: %w", err)
		}
		for _, v := range req.Variants {
			_, err = tx.Exec(ctx, `
				INSERT INTO product_variants (product_id, color, size, stock_quantity)
				VALUES ($1, $2, $3, $4)
			`, p.ID, v.Color, v.Size, v.StockQuantity)
			if err != nil {
				return nil, fmt.Errorf("failed to insert variant: %w", err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	p.Images, _ = r.GetProductImages(ctx, p.ID)
	p.Variants, _ = r.GetProductVariants(ctx, p.ID)

	return &p, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx,
		"DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// ─── Images ──────────────────────────────────────────────────

func (r *ProductRepository) GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, url, COALESCE(public_id, ''), sort_order, created_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY sort_order ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		err := rows.Scan(
			&img.ID, &img.ProductID, &img.URL,
			&img.PublicID, &img.SortOrder, &img.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

// ─── Variants ────────────────────────────────────────────────

func (r *ProductRepository) GetProductVariants(ctx context.Context, productID string) ([]models.ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, color, size, stock_quantity, created_at
		FROM product_variants
		WHERE product_id = $1
		ORDER BY color, size
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []models.ProductVariant
	for rows.Next() {
		var v models.ProductVariant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.Color,
			&v.Size, &v.StockQuantity, &v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		variants = append(variants, v)
	}
	return variants, nil
}

func (r *ProductRepository) GetVariantByID(ctx context.Context, id string) (*models.ProductVariant, error) {
	var v models.ProductVariant
	err := r.db.QueryRow(ctx, `
		SELECT id, product_id, color, size, stock_quantity, created_at
		FROM product_variants WHERE id = $1
	`, id).Scan(
		&v.ID, &v.ProductID, &v.Color,
		&v.Size, &v.StockQuantity, &v.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("variant not found: %w", err)
	}
	return &v, nil
}

// ─── Categories ──────────────────────────────────────────────

func (r *ProductRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, name, COALESCE(description, ''), created_at FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *ProductRepository) CreateCategory(ctx context.Context, name, description string, cat *models.Category) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, COALESCE(description, ''), created_at
	`, name, description,
	).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
}

func (r *ProductRepository) DeleteCategory(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx,
		"DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}