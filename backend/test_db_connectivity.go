package main

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/lib/pq"
)

func main() {
	urls := map[string]string{
		"auth": os.Getenv("DEV_CENTRAL_AUTH_DB_URL"),
		"longbeach": os.Getenv("DEV_LONGBEACH_DB_URL"), 
		"bakersfield": os.Getenv("DEV_BAKERSFIELD_DB_URL"),
		"colorado": os.Getenv("DEV_COLORADO_DB_URL"),
	}
	
	for name, url := range urls {
		if url == "" {
			fmt.Printf("❌ %s: URL not set\n", name)
			continue
		}
		
		db, err := sql.Open("postgres", url)
		if err != nil {
			fmt.Printf("❌ %s: Failed to open - %v\n", name, err)
			continue
		}
		
		if err := db.Ping(); err != nil {
			fmt.Printf("❌ %s: Failed to ping - %v\n", name, err)
			db.Close()
			continue
		}
		
		fmt.Printf("✅ %s: Connected successfully\n", name)
		db.Close()
	}
}