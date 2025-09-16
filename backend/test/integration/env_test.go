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
    requiredVars := []string{"CENTRAL_AUTH_DB_URL", "LONGBEACH_DB_URL"}
    
    for _, varName := range requiredVars {
        value := os.Getenv(varName)
        if value == "" {
            t.Errorf("Required environment variable %s is not set", varName)
        } else {
            t.Logf("✅ %s is set", varName)
        }
    }
    
    // Test database URL formats
    centralAuthURL := os.Getenv("CENTRAL_AUTH_DB_URL")
    longbeachURL := os.Getenv("LONGBEACH_DB_URL")
    
    if centralAuthURL != "" {
        if len(centralAuthURL) < 10 {
            t.Errorf("CENTRAL_AUTH_DB_URL seems too short: %s", centralAuthURL)
        } else {
            t.Logf("✅ CENTRAL_AUTH_DB_URL format looks good")
        }
    }
    
    if longbeachURL != "" {
        if len(longbeachURL) < 10 {
            t.Errorf("LONGBEACH_DB_URL seems too short: %s", longbeachURL)
        } else {
            t.Logf("✅ LONGBEACH_DB_URL format looks good")
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
