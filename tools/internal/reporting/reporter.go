package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"oilgas-tools/internal/processor"
)

// Reporter handles report generation
type Reporter struct {
	outputDir string
}

// New creates a new reporter
func New(outputDir string) *Reporter {
	return &Reporter{outputDir: outputDir}
}

// ProcessingReport represents a processing report
type ProcessingReport struct {
	GeneratedAt time.Time                   `json:"generated_at"`
	Company     string                      `json:"company"`
	InputFile   string                      `json:"input_file"`
	Summary     ProcessingSummary           `json:"summary"`
	OutputFiles []string                    `json:"output_files"`
}

// ProcessingSummary provides statistics
type ProcessingSummary struct {
	TotalRecords   int     `json:"total_records"`
	ValidRecords   int     `json:"valid_records"`
	ErrorRecords   int     `json:"error_records"`
	SuccessRate    float64 `json:"success_rate_percent"`
	ProcessingTime string  `json:"processing_time"`
}

// GenerateReport creates a processing report
func (r *Reporter) GenerateReport(company, inputFile string, result *processor.ProcessingResult) error {
	successRate := float64(0)
	if result.RecordsProcessed > 0 {
		successRate = float64(result.ValidRecords) / float64(result.RecordsProcessed) * 100
	}

	report := &ProcessingReport{
		GeneratedAt: time.Now(),
		Company:     company,
		InputFile:   inputFile,
		Summary: ProcessingSummary{
			TotalRecords:   result.RecordsProcessed,
			ValidRecords:   result.ValidRecords,
			ErrorRecords:   result.Errors,
			SuccessRate:    successRate,
			ProcessingTime: result.Duration.String(),
		},
		OutputFiles: result.OutputFiles,
	}

	// Write JSON report
	reportsDir := filepath.Join(r.outputDir, "reports")
	os.MkdirAll(reportsDir, 0755)

	filename := fmt.Sprintf("report_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(reportsDir, filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}
