//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Get database URL from command line argument or environment
	dbURL := os.Getenv("DATABASE_URL")
	if len(os.Args) > 1 {
		dbURL = os.Args[1]
	}
	
	if dbURL == "" {
		fmt.Println("❌ No database URL provided")
		os.Exit(1)
	}
	
	fmt.Printf("Testing connection to: %s\n", maskPassword(dbURL))
	
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	
	err = pool.Ping(context.Background())
	if err != nil {
		fmt.Printf("❌ Ping failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Database connection successful")
}

func maskPassword(url string) string {
	// Simple password masking for logs
	return url[:strings.Index(url, "://")+3] + "***:***@" + url[strings.LastIndex(url, "@")+1:]
}
