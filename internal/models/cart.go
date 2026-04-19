package models

type CartItem struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
	ImageURL     string  `json:"image_url"`
	Color        string  `json:"color"`
	Size         string  `json:"size"`
	Quantity     int     `json:"quantity"`
	Subtotal     float64 `json:"subtotal"`
}

type Cart struct {
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"total_amount"`
	TotalItems  int        `json:"total_items"`
}