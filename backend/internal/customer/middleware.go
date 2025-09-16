// backend/internal/customer/middleware.go
package customer

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
)

// TenantIsolationMiddleware ensures proper tenant isolation
func TenantIsolationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			tenantID = c.Query("tenant_id")
		}
		
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID required"})
			c.Abort()
			return
		}
		
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

// PerformanceMonitoringMiddleware tracks API performance
func PerformanceMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		// Log slow requests (in production, use proper metrics)
		if duration > time.Second*2 {
			fmt.Printf("SLOW REQUEST: %s %s took %v\n", 
				c.Request.Method, c.Request.URL.Path, duration)
		}
		
		// Set response headers for debugging
		c.Header("X-Response-Time", duration.String())
	}
}

// CustomerAccessMiddleware ensures customer contacts can only access their own data
func CustomerAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from auth middleware
		userRole := c.GetString("user_role")
		customerFilter := c.GetInt("customer_filter")
		
		// If user is a customer contact, validate access
		if userRole == "customer_contact" && customerFilter > 0 {
			// Check if requested customer ID matches user's customer
			if customerIDParam := c.Param("id"); customerIDParam != "" {
				requestedCustomerID, err := strconv.Atoi(customerIDParam)
				if err != nil || requestedCustomerID != customerFilter {
					c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
					c.Abort()
					return
				}
			}
		}
		
		c.Next()
	}
}

// RateLimitingMiddleware provides basic rate limiting per tenant
func RateLimitingMiddleware() gin.HandlerFunc {
	// In production, use Redis or proper rate limiting service
	limits := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			c.Next()
			return
		}
		
		now := time.Now()
		windowStart := now.Add(-time.Minute)
		
		// Clean old entries
		var recentRequests []time.Time
		for _, reqTime := range limits[tenantID] {
			if reqTime.After(windowStart) {
				recentRequests = append(recentRequests, reqTime)
			}
		}
		
		// Check rate limit (100 requests per minute per tenant)
		if len(recentRequests) >= 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": "60s",
			})
			c.Abort()
			return
		}
		
		// Add current request
		recentRequests = append(recentRequests, now)
		limits[tenantID] = recentRequests
		
		c.Next()
	}
}
