package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"tshirt-store/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

// POST /api/products/createProduct (Protected)
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	log.Println("📥 [PRODUCT-HANDLER] Processing detailed apparel creation request...")

	dto := new(models.CreateProductDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload format"})
	}

	// Basic validation
	if dto.Name == "" || dto.ProductCode == "" || dto.SKU == "" || dto.CategoryID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, ProductCode, SKU, and CategoryID are strictly required"})
	}

	// 1. Calculate discount percentage automatically
	var discountPct int
	if dto.MRP > 0 && dto.SellingPrice < dto.MRP {
		discountPct = int(((dto.MRP - dto.SellingPrice) / dto.MRP) * 100)
	}

	// 2. Convert Images slice to JSON string
	imagesJSON, _ := json.Marshal(dto.Images)

	// 3. Convert Specifications map to JSON string
	specsJSON, _ := json.Marshal(dto.Specifications)

	// 4. Build Product entity
	product := models.Product{
		Name:           dto.Name,
		Brand:          dto.Brand,
		Description:    dto.Description,
		ProductCode:    dto.ProductCode,
		SKU:            dto.SKU,
		MRP:            dto.MRP,
		SellingPrice:   dto.SellingPrice,
		Discount:       discountPct,
		CategoryID:     dto.CategoryID,
		Images:         string(imagesJSON),
		Specifications: string(specsJSON),
	}

	// 5. Build Size variants
	for _, sizeDTO := range dto.Sizes {
		product.Sizes = append(product.Sizes, models.ProductSize{
			Size:  sizeDTO.Size,
			Stock: sizeDTO.Stock,
		})
	}

	// 6. Save cleanly in PostgreSQL (GORM will save Product and ProductSizes in a single transaction)
	if err := h.DB.Create(&product).Error; err != nil {
		log.Printf("🚨 [PRODUCT-HANDLER] DB Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create product record"})
	}

	log.Printf("🎉 [PRODUCT-HANDLER] Success! Apparel '%s' (Code: %s) created with ID #%d", product.Name, product.ProductCode, product.ID)
	return c.JSON(http.StatusCreated, product)
}

// GET /api/products/getAllProducts (Public)
func (h *ProductHandler) GetAllProducts(c echo.Context) error {
	var products []models.Product
	// Preload Category and Sizes breakdown
	if err := h.DB.Preload("Category").Preload("Sizes").Find(&products).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch catalog products"})
	}
	return c.JSON(http.StatusOK, products)
}

// GET /api/products/updateProduct/:id (Protected)
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	idParam := c.Param("id")
	productID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID format"})
	}
	var existingProduct models.Product
	if err := h.DB.Preload("Sizes").First(&existingProduct, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database lookup failed"})
	}

	dto := new(models.UpdateProductDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload format"})
	}

	// Basic validation
	// if dto.Name == "" || dto.ProductCode == "" || dto.SKU == "" || dto.CategoryID == 0 {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, ProductCode, SKU, and CategoryID are strictly required"})
	// }
	updates := make(map[string]interface{})

	if dto.Name != "" {
		updates["name"] = dto.Name
	}
	if dto.Brand != "" {
		updates["brand"] = dto.Brand
	}
	if dto.Description != "" {
		updates["description"] = dto.Description
	}
	if dto.ProductCode != "" {
		updates["product_code"] = dto.ProductCode
	}
	if dto.SKU != "" {
		updates["sku"] = dto.SKU
	}
	mrp := existingProduct.MRP
	sellingPrice := existingProduct.SellingPrice
	if dto.MRP > 0 {
		mrp = dto.MRP
		updates["mrp"] = dto.MRP
	}
	if dto.SellingPrice > 0 {
		sellingPrice = dto.SellingPrice
		updates["selling_price"] = dto.SellingPrice
	}

	if mrp > 0 && sellingPrice < mrp {
		updates["discount"] = int(((mrp - sellingPrice) / mrp) * 100)
	}
	// Convert updated images slice to JSON string if provided
	if len(dto.Images) > 0 {
		imagesJSON, _ := json.Marshal(dto.Images)
		updates["images"] = string(imagesJSON)
	}

	// Convert updated specifications map to JSON string if provided
	if len(dto.Specifications) > 0 {
		specsJSON, _ := json.Marshal(dto.Specifications)
		updates["specifications"] = string(specsJSON)
	}
	// 4. Begin DB Transaction for Atomic Update
	tx := h.DB.Begin()
	// A. Execute SQL UPDATE on the primary products row
	if len(updates) > 0 {
		if err := tx.Model(&existingProduct).Updates(updates).Error; err != nil {
			tx.Rollback()
			log.Printf("🚨 [PRODUCT-HANDLER] DB Update error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update product attributes"})
		}
	}
	if len(dto.Sizes) > 0 {
		// Permanently clear existing size variants for this product
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&models.ProductSize{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to refresh size variants"})
		}

		// Insert fresh size records with updated stock numbers
		var newSizes []models.ProductSize
		for _, sizeDTO := range dto.Sizes {
			newSizes = append(newSizes, models.ProductSize{
				ProductID: uint(productID),
				Size:      sizeDTO.Size,
				Stock:     sizeDTO.Stock,
			})
		}

		if err := tx.Create(&newSizes).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create updated size variants"})
		}
	}

	// Commit Transaction
	tx.Commit()

	// 5. Fetch and return the freshly updated product record
	var updatedProduct models.Product
	h.DB.Preload("Category").Preload("Sizes").First(&updatedProduct, productID)

	log.Printf("✏️ [PRODUCT-HANDLER] Product ID #%d ('%s') successfully updated in DB", productID, updatedProduct.Name)
	return c.JSON(http.StatusOK, updatedProduct)

}

// DELETE /api/products/deleteProduct/:id (Protected)
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	idParam := c.Param("id")
	productID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID format"})
	}

	// 1. Check if product exists (include Unscoped so we find it regardless)
	var product models.Product
	if err := h.DB.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database lookup failed"})
	}

	// 2. Use Unscoped() + Transaction to permanently delete Product AND its Size variants
	tx := h.DB.Begin()

	// Delete child size variants permanently
	if err := tx.Unscoped().Where("product_id = ?", product.ID).Delete(&models.ProductSize{}).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete size inventory"})
	}

	// Delete primary product permanently
	if err := tx.Unscoped().Delete(&product).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete product record"})
	}

	tx.Commit()

	log.Printf("💥 [PRODUCT-HANDLER] Product ID #%d and its variants PERMANENTLY deleted from DB", productID)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Product permanently deleted from database",
	})
}
