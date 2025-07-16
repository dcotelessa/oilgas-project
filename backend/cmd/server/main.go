package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Set Gin mode
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "oil-gas-inventory-api",
			"version":   "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Oil & Gas Inventory API",
				"status":  "running",
				"env":     os.Getenv("APP_ENV"),
			})
		})

		// Placeholder endpoints for future development
		v1.GET("/customers", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Customers endpoint - coming soon"})
		})

		v1.GET("/inventory", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Inventory endpoint - coming soon"})
		})

		v1.GET("/received", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Received endpoint - coming soon"})
		})
	}

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("🚀 Starting Oil & Gas Inventory API server on port %s\n", port)
	fmt.Printf("📋 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔌 API base: http://localhost:%s/api/v1\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
