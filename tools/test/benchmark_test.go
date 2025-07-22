/ tools/test/benchmark_test.go
package test

import (
	"testing"
	"oilgas-tools/internal/mapping"
	"oilgas-tools/internal/processor"
)

func BenchmarkGradeNormalization(b *testing.B) {
	grades := []string{"J-55", "L80", "P-110", "N80", "Q125"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grade := grades[i%len(grades)]
		mapping.NormalizeGrade(grade)
	}
}

func BenchmarkTableProcessing(b *testing.B) {
	// Create sample data
	sampleData := make([][]string, 1000)
	for i := range sampleData {
		sampleData[i] = []string{
			fmt.Sprintf("LB-%06d", i),
			"Test Customer",
			"100",
			"5 1/2",
			"L80",
		}
	}

	tableInfo := &processor.TableInfo{
		Name:       "test_table",
		RowCount:   1000,
		Columns:    []string{"work_order", "customer", "joints", "size", "grade"},
		DataSample: sampleData,
	}

	job := &processor.ConversionJob{
		Config: &processor.Config{
			Workers:   1,
			BatchSize: 100,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job.ProcessTable(context.Background(), tableInfo)
	}
}

// Property-based testing example
func TestColumnMapping_Properties(t *testing.T) {
	// Property: Normalization should be idempotent
	t.Run("grade_normalization_idempotent", func(t *testing.T) {
		testCases := []string{"J55", "L80", "P110", "N80", "Q125"}
		
		for _, grade := range testCases {
			normalized1, valid1 := mapping.NormalizeGrade(grade)
			if !valid1 {
				continue
			}
			
			normalized2, valid2 := mapping.NormalizeGrade(normalized1)
			
			assert.True(t, valid2)
			assert.Equal(t, normalized1, normalized2, 
				"Normalization should be idempotent for %s", grade)
		}
	})

	// Property: Valid inputs should always produce valid outputs
	t.Run("size_validation_consistency", func(t *testing.T) {
		validSizes := []string{"4 1/2\"", "5\"", "5 1/2\"", "7\"", "9 5/8\""}
		
		for _, size := range validSizes {
			normalized, valid := mapping.NormalizeSize(size)
			
			assert.True(t, valid, "Valid size %s should remain valid", size)
			assert.NotEmpty(t, normalized, "Valid size should produce non-empty result")
		}
	})
}
