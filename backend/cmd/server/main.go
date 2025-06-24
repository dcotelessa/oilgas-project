package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from project root or local
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found")
		}
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}


	// Initialize Gin router
	r := gin.Default()

	// Configure trusted proxies (fix the warning)
	// In development, trust localhost and docker networks
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// CORS middleware
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// In development, allow localhost origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "oil-gas-inventory",
			"version": "1.0.0",
			"environment": os.Getenv("APP_ENV"),
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/grades", getGrades)
		api.GET("/customers", getCustomers)
		api.GET("/inventory", getInventory)
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}

// Placeholder handlers
func getGrades(c *gin.Context) {
	grades := []string{"J55", "JZ55", "L80", "N80", "P105", "P110"}
	c.JSON(http.StatusOK, gin.H{"grades": grades})
}

func getCustomers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"customers": []interface{}{}})
}

func getInventory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"inventory": []interface{}{}})
}
