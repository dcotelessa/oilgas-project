package processor

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"oilgas-tools/internal/mapping"
)

// Processor handles the main processing logic
type Processor struct {
	config *Config
	mapper *mapping.ColumnMapper
}

// New creates a new processor instance
func New(config *Config) *Processor {
	return &Processor{
		config: config,
		mapper: mapping.NewColumnMapper(config.MappingConfig),
	}
}

// ProcessFile processes a single file
func (p *Processor) ProcessFile(ctx context.Context, filename string) (*ProcessingResult, error) {
	start := time.Now()

	// Create output directories
	if err := p.createOutputDirs(); err != nil {
		return nil, fmt.Errorf("failed to create output directories: %w", err)
	}

	// Analyze file
	tableInfo, err := p.analyzeFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze file: %w", err)
	}

	if p.config.Verbose {
		fmt.Printf("📊 Processing: %s (%d rows, %d columns)\n", 
			filename, tableInfo.RowCount, len(tableInfo.Columns))
	}

	// Process the file
	result, err := p.processTable(ctx, tableInfo, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to process table: %w", err)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// analyzeFile examines the input file
func (p *Processor) analyzeFile(filename string) (*TableInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Count rows
	rowCount := 0
	sampleData := make([][]string, 0, 3)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rowCount++
		if len(sampleData) < 3 {
			sampleData = append(sampleData, record)
		}
	}

	tableName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	return &TableInfo{
		Name:       tableName,
		RowCount:   rowCount,
		Columns:    headers,
		DataSample: sampleData,
		FileSize:   stat.Size(),
	}, nil
}

// processTable processes a single table
func (p *Processor) processTable(ctx context.Context, tableInfo *TableInfo, filename string) (*ProcessingResult, error) {
	result := &ProcessingResult{
		ValidationIssues: make([]ValidationIssue, 0),
		OutputFiles:      make([]string, 0),
	}

	// Open input file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Read headers
	originalHeaders, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Normalize headers
	normalizedHeaders := p.mapper.NormalizeHeaders(originalHeaders)

	// Create output files
	csvOutput := filepath.Join(p.config.OutputDir, "csv", tableInfo.Name+".csv")
	sqlOutput := filepath.Join(p.config.OutputDir, "sql", tableInfo.Name+".sql")

	csvFile, err := os.Create(csvOutput)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	sqlFile, err := os.Create(sqlOutput)
	if err != nil {
		return nil, err
	}
	defer sqlFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Write headers
	csvWriter.Write(normalizedHeaders)

	// Write SQL header
	fmt.Fprintf(sqlFile, "-- Generated for %s\n", p.config.Company)
	fmt.Fprintf(sqlFile, "-- Table: %s\n", tableInfo.Name)
	fmt.Fprintf(sqlFile, "COPY store.%s FROM '%s' WITH CSV HEADER;\n", 
		tableInfo.Name, csvOutput)

	// Process rows
	rowNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors++
			continue
		}

		rowNum++
		result.RecordsProcessed++

		// Pad or trim record to match headers
		for len(record) < len(normalizedHeaders) {
			record = append(record, "")
		}
		if len(record) > len(normalizedHeaders) {
			record = record[:len(normalizedHeaders)]
		}

		// Apply basic transformations
		transformedRecord := p.processRecord(record, normalizedHeaders)
		result.ValidRecords++

		// Write to CSV
		csvWriter.Write(transformedRecord)
	}

	result.OutputFiles = append(result.OutputFiles, csvOutput, sqlOutput)

	if p.config.Verbose {
		fmt.Printf("✅ Processed %d records\n", result.RecordsProcessed)
	}

	return result, nil
}

// processRecord applies transformations to a record
func (p *Processor) processRecord(record []string, headers []string) []string {
	transformed := make([]string, len(record))
	copy(transformed, record)
	
	for i, value := range record {
		if i >= len(headers) {
			continue
		}
		
		columnName := strings.ToLower(headers[i])
		
		// Apply business rules
		if strings.Contains(columnName, "grade") {
			if norm, valid := p.mapper.NormalizeGrade(value); valid {
				transformed[i] = norm
			}
		} else if strings.Contains(columnName, "size") {
			if norm, valid := p.mapper.NormalizeSize(value); valid {
				transformed[i] = norm
			}
		} else if strings.Contains(columnName, "customer") {
			transformed[i] = p.mapper.NormalizeCustomerName(value)
		}
	}
	
	return transformed
}

// createOutputDirs creates output directories
func (p *Processor) createOutputDirs() error {
	dirs := []string{
		filepath.Join(p.config.OutputDir, "csv"),
		filepath.Join(p.config.OutputDir, "sql"),
		filepath.Join(p.config.OutputDir, "reports"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
