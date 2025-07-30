// backend/cmd/tools/tenant_processor/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: tenant_processor <directory>")
	}

	directory := os.Args[1]
	err := ProcessDirectory(directory)
	if err != nil {
		log.Fatal(err)
	}
}

func ProcessDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".csv") {
			fmt.Printf("Processing: %s\n", path)
			// TODO: Implement tenant-specific CSV processing logic
		}
		
		return nil
	})
}

func NormalizeColumnName(column string) string {
	normalized := strings.ToLower(strings.TrimSpace(column))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, "/", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")
	return normalized
}
