package integration_test

import (
    "os"
    "path/filepath"
    "testing"
    
    "github.com/joho/godotenv"
)

func TestEnvironmentLoads(t *testing.T) {
    // Find .env.local file - try multiple possible locations
    envPaths := []string{
        "../../../.env.local",  // From backend/test/integration/ -> root
        "../../.env.local",     // From backend/test/ -> root  
        ".env.local",           // Current directory
        "../.env.local",        // One level up
    }
    
    var envFile string
    var err error
    
    for _, path := range envPaths {
        if _, err := os.Stat(path); err == nil {
            envFile = path
            break
        }
    }
    
    if envFile == "" {
        t.Fatal("Could not find .env.local file. Tried paths: ../../../.env.local, ../../.env.local, .env.local, ../.env.local")
    }
    
    t.Logf("Found .env.local at: %s", envFile)
    
    // Test that .env.local can be loaded
    err = godotenv.Load(envFile)
    if err != nil {
        t.Fatalf("Failed to load %s: %v", envFile, err)
    }
    
    // Test key environment variables exist
    requiredVars := []string{"DATABASE_URL", "POSTGRES_DB"}
    
    for _, varName := range requiredVars {
        value := os.Getenv(varName)
        if value == "" {
            t.Errorf("Required environment variable %s is not set", varName)
        } else {
            t.Logf("✅ %s is set", varName)
        }
    }
    
    // Test DATABASE_URL format
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL != "" {
        if len(databaseURL) < 10 {
            t.Errorf("DATABASE_URL seems too short: %s", databaseURL)
        } else {
            t.Logf("✅ DATABASE_URL format looks good")
        }
    }
}

func TestWorkingDirectory(t *testing.T) {
    // Show current working directory for debugging
    wd, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get working directory: %v", err)
    }
    
    t.Logf("Current working directory: %s", wd)
    
    // Show what files exist in potential .env.local locations
    envPaths := []string{
        "../../../.env.local",
        "../../.env.local", 
        ".env.local",
        "../.env.local",
    }
    
    for _, path := range envPaths {
        absPath, _ := filepath.Abs(path)
        if _, err := os.Stat(path); err == nil {
            t.Logf("✅ Found: %s (absolute: %s)", path, absPath)
        } else {
            t.Logf("❌ Not found: %s (absolute: %s)", path, absPath)
        }
    }
}
