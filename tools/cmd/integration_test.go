// tools/cmd/integration_test.go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestMDBProcessor_Integration(t *testing.T) {
	// Skip if mdb_processor binary doesn't exist
	binaryPath := filepath.Join(".", "bin", "mdb_processor")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("mdb_processor binary not found, run 'make tools-build' first")
	}

	t.Run("help_command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-help")
		output, err := cmd.CombinedOutput()
		
		assert.NoError(t, err)
		assert.Contains(t, string(output), "Usage:")
		assert.Contains(t, string(output), "Options:")
	})

	t.Run("config_validation", func(t *testing.T) {
		// Test with invalid config
		cmd := exec.Command(binaryPath, "-config", "nonexistent.json")
		err := cmd.Run()
		
		assert.Error(t, err)
		// Should exit with non-zero status for invalid config
	})

	t.Run("sample_conversion", func(t *testing.T) {
		// Create temporary test file
		testFile := filepath.Join(os.TempDir(), "test_data.csv")
		testData := `WorkOrder,Customer,Joints,Size,Grade
LB-001001,Test Customer,100,5 1/2,L80
LB-001002,Another Customer,150,7,P110`
		
		err := os.WriteFile(testFile, []byte(testData), 0644)
		assert.NoError(t, err)
		defer os.Remove(testFile)

		// Create temporary output directory
		outputDir := filepath.Join(os.TempDir(), "test_output")
		os.MkdirAll(outputDir, 0755)
		defer os.RemoveAll(outputDir)

		// Run conversion
		cmd := exec.Command(binaryPath,
			"-file", testFile,
			"-company", "Test Company",
			"-output", outputDir,
		)
		
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Command output: %s", string(output))

		// Verify outputs were created
		csvOutput := filepath.Join(outputDir, "csv")
		assert.DirExists(t, csvOutput)
		
		sqlOutput := filepath.Join(outputDir, "sql")
		assert.DirExists(t, sqlOutput)
	})
}
