package main

import (
	"log"

	"tshirt-store/internal/config"
	"tshirt-store/internal/db"
	"tshirt-store/internal/handlers"
	"tshirt-store/internal/models"
	"tshirt-store/internal/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	log.Println("🚀 [MAIN] Bootstrapping backend server lifecycle...")

	// 1. Ingest environment configurations
	cfg := config.Load()

	// 2. Initialize the GORM framework connection pool
	database := db.ConnectORM(cfg.DatabaseURL)

	// 3. Sync structural tables via ORM AutoMigration
	log.Println("🔄 [MIGRATION] Scanning entity mappings for auto-migration...")
	err := database.AutoMigrate(&models.User{}, &models.Product{}, &models.ProductSize{}, &models.Category{})
	if err != nil {
		log.Fatalf("🚨 [MIGRATION] Schema sync failed! System shutting down. Trace: %v", err)
	}
	log.Println("✅ [MIGRATION] Target database tables synchronized.")

	// 4. Spin up the Echo framework router
	log.Println("🤖 [ROUTER] Assembling Echo framework engine...")
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			cfg.FrontendURL,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowMethods: []string{
			echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS,
		},
	}))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	log.Println("🛡️  [ROUTER] Global system middleware layers attached.")

	// 5. Instantiate your route handlers
	log.Println("📦 [MAIN] Instantiating route handlers and injecting dependencies...")
	authHandler := &handlers.AuthHandler{
		DB:        database,
		JWTSecret: []byte(cfg.JWTSecret),
	}
	productHandler := &handlers.ProductHandler{DB: database}
	categoryHandler := &handlers.CategoryHandler{DB: database}

	// 6. 🔥 CALL THE ROUTE MAPPER (This completely replaces the old route code)
	routes.SetupRoutes(e, authHandler, productHandler, categoryHandler, []byte(cfg.JWTSecret))

	// 7. Bind listener loops onto target network port parameters
	log.Printf("🏁 [MAIN] System baseline startup complete. Opening listeners at address: 0.0.0.0:%s", cfg.Port)
	if err := e.Start("0.0.0.0:" + cfg.Port); err != nil {
		log.Fatalf("🚨 [MAIN] Network server binding encountered a critical error loop: %v", err)
	}
}
