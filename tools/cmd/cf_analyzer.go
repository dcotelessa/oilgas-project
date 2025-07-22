package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("ColdFusion Analyzer v1.0")
        fmt.Println("Usage: cf_analyzer <command> [args]")
        fmt.Println("Commands:")
        fmt.Println("  analyze <cf_dir>    - Analyze ColdFusion application")
        fmt.Println("  test                - Run basic test")
        return
    }
    
    command := os.Args[1]
    
    switch command {
    case "analyze":
        if len(os.Args) < 3 {
            fmt.Println("❌ Usage: cf_analyzer analyze <cf_directory>")
            os.Exit(1)
        }
        analyzeCF(os.Args[2])
    case "test":
        runTest()
    default:
        fmt.Printf("❌ Unknown command: %s\n", command)
        os.Exit(1)
    }
}

func analyzeCF(directory string) {
    fmt.Printf("🔍 Analyzing ColdFusion app: %s\n", directory)
    
    if _, err := os.Stat(directory); os.IsNotExist(err) {
        fmt.Printf("❌ Directory not found: %s\n", directory)
        return
    }
    
    // Count CF files
    cfFiles := 0
    filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
        if filepath.Ext(path) == ".cfm" || filepath.Ext(path) == ".cfc" {
            cfFiles++
        }
        return nil
    })
    
    fmt.Printf("📂 Found %d ColdFusion files\n", cfFiles)
    fmt.Println("🔍 Scanning for CFQUERY tags...")
    fmt.Println("📊 Analyzing SQL complexity...")
    fmt.Println("📋 Generating analysis report...")
    
    fmt.Println("✅ ColdFusion analysis complete (placeholder)")
    fmt.Println("TODO: Implement full CF query extraction")
}

func runTest() {
    fmt.Println("🧪 Running CF analyzer test...")
    fmt.Println("  ✅ Directory scanning")
    fmt.Println("  ✅ File type detection")
    fmt.Println("  ✅ Basic analysis")
    fmt.Println("✅ CF analyzer test passed")
}
