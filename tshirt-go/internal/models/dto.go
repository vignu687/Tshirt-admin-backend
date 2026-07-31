package models

type SizeStockDTO struct {
	Size  string `json:"size"`  // "S", "M", "L", "XL"
	Stock int    `json:"stock"` // Quantity available
}

type CreateProductDTO struct {
	Name           string            `json:"name"`
	Brand          string            `json:"brand"`
	Description    string            `json:"description"`
	ProductCode    string            `json:"product_code"`
	SKU            string            `json:"sku"`
	MRP            float64           `json:"mrp"`
	SellingPrice   float64           `json:"selling_price"`
	CategoryID     uint              `json:"category_id"`
	Images         []string          `json:"images"`         // Array of image URLs
	Specifications map[string]string `json:"specifications"` // Dynamic attributes (Fabric, Fit, etc.)
	Sizes          []SizeStockDTO    `json:"sizes"`          // Size inventory breakdown
}
type UpdateProductDTO struct {
	Name           string            `json:"name"`
	Brand          string            `json:"brand"`
	Description    string            `json:"description"`
	ProductCode    string            `json:"product_code"`
	SKU            string            `json:"sku"`
	MRP            float64           `json:"mrp"`
	SellingPrice   float64           `json:"selling_price"`
	CategoryID     uint              `json:"category_id"`
	Images         []string          `json:"images"`
	Specifications map[string]string `json:"specifications"`
	Sizes          []SizeStockDTO    `json:"sizes"`
}
