package exporters

import (
	"encoding/csv"
	"io"
)

// CSVExporter handles CSV exports
type CSVExporter struct{}

// NewCSVExporter creates a CSV exporter
func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

// Export writes data to CSV
func (e *CSVExporter) Export(data [][]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()
	
	for _, record := range data {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	return nil
}
