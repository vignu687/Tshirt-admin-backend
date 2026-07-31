package routes

import (
	"log"
	"net/http"

	"tshirt-store/internal/handlers"
	"tshirt-store/internal/middleware" // Import your custom middleware package

	"github.com/labstack/echo/v4"
)

// SetupRoutes acts as the master map for all API endpoints in the application
func SetupRoutes(
	e *echo.Echo,
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	categoryHandler *handlers.CategoryHandler,
	jwtSecret []byte, // Injected secret key for token verification
) {
	log.Println("🗺️  [ROUTES] Registering application route mappings...")

	// 🔑 Instantiate the custom authentication middleware
	authRequired := middleware.JWTMiddleware(jwtSecret)

	// 1. Core System Infrastructure Endpoints
	e.GET("/health", func(c echo.Context) error {
		log.Println("🔍 [ROUTES] Health check endpoint touched.")
		return c.JSON(http.StatusOK, map[string]string{
			"status": "online",
			"orm":    "synchronized",
		})
	})

	// 2. Authentication Route Group (Public)
	authGroup := e.Group("/api/auth")
	{
		log.Println("🔒 [ROUTES] Binding authentication subsystem routes...")
		authGroup.POST("/signup", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// 3. Category Routing Blueprint
	categoryGroup := e.Group("/api/categories")
	{
		log.Println("📁 [ROUTES] Binding category configuration routes...")
		// 🌐 Public: Anyone can view categories
		categoryGroup.GET("/getAllCategories", categoryHandler.GetAllCategories)

		// 🔒 Protected: Requires a valid Bearer token signature
		categoryGroup.POST("/createCategory", categoryHandler.CreateCategory, authRequired)
		categoryGroup.PUT("/updateCategory/:id", categoryHandler.UpdateCategory, authRequired)
		categoryGroup.DELETE("/deleteCategory/:id", categoryHandler.DeleteCategory, authRequired)
	}

	// 4. Product Routing Blueprint
	productGroup := e.Group("/api/products")
	{
		log.Println("👕 [ROUTES] Binding product core routes...")
		// 🌐 Public: Anyone can view products
		productGroup.GET("/getAllProducts", productHandler.GetAllProducts)

		// 🔒 Protected: Requires a valid Bearer token signature
		productGroup.POST("/createProduct", productHandler.CreateProduct, authRequired)
		productGroup.PUT("/updateProduct/:id", productHandler.UpdateProduct, authRequired)
		productGroup.DELETE("/deleteProduct/:id", productHandler.DeleteProduct, authRequired)
	}

	log.Println("✅ [ROUTES] All backend network endpoints mapped successfully.")
}
