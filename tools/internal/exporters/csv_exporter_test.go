// tools/internal/exporters/csv_exporter_test.go
package exporters

import (
	"bytes"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCSVExporter_Export(t *testing.T) {
	exporter := NewCSVExporter()
	
	t.Run("basic_export", func(t *testing.T) {
		data := [][]string{
			{"work_order", "customer", "joints"},
			{"LB-001001", "Test Customer", "100"},
			{"LB-001002", "Another Customer", "150"},
		}

		var buf bytes.Buffer
		err := exporter.Export(data, &buf)
		
		assert.NoError(t, err)
		
		output := buf.String()
		assert.Contains(t, output, "work_order,customer,joints")
		assert.Contains(t, output, "LB-001001,Test Customer,100")
		assert.Contains(t, output, "LB-001002,Another Customer,150")
	})

	t.Run("special_characters", func(t *testing.T) {
		data := [][]string{
			{"customer", "notes"},
			{"Customer with, comma", "Notes with \"quotes\""},
		}

		var buf bytes.Buffer
		err := exporter.Export(data, &buf)
		
		assert.NoError(t, err)
		
		output := buf.String()
		// Verify proper CSV escaping
		assert.Contains(t, output, "\"Customer with, comma\"")
		assert.Contains(t, output, "\"Notes with \"\"quotes\"\"\"")
	})
}
