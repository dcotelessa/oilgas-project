// internal/processor/processor.go
// Core MDB processing logic
package processor

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"tools/internal/config"
	"tools/internal/mapping"
	"tools/internal/validation"
	"tools/internal/exporters"
	_ "github.com/lib/pq"
)

// Processor handles MDB file processing operations
type Processor struct {
	config    *config.Config
	db        *sql.DB
	mapper    *mapping.ColumnMapper
	validator *validation.Validator
	exporters map[string]exporters.Exporter
	
	// Performance monitoring
	performanceMonitor *PerformanceMonitor
	
	// Worker management
	workerPool     *WorkerPool
	jobQueue       chan WorkerJob
	resultQueue    chan WorkerResult
	
	// State management
	currentJob     *ConversionJob
	mutex          sync.RWMutex
}

// New creates a new MDB processor instance
func New(cfg *config.Config, dbConnString string) (*Processor, error) {
	log.Println("Initializing MDB processor...")
	
	// Initialize database connection
	db, err := initializeDatabase(dbConnString, cfg.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	
	// Initialize components
	mapper := mapping.New(cfg)
	validator := validation.New(cfg)
	
	// Initialize exporters
	exporterMap := make(map[string]exporters.Exporter)
	if cfg.OutputSettings.CSVOutput {
		exporterMap["csv"] = exporters.NewCSVExporter(cfg)
	}
	if cfg.OutputSettings.SQLOutput {
		exporterMap["sql"] = exporters.NewSQLExporter(cfg)
	}
	if cfg.OutputSettings.PostgreSQLDirect {
		exporterMap["direct"] = exporters.NewDirectExporter(cfg, db)
	}
	
	// Initialize worker pool
	workerPool := NewWorkerPool(cfg.ProcessingOptions.Workers)
	
	processor := &Processor{
		config:             cfg,
		db:                 db,
		mapper:             mapper,
		validator:          validator,
		exporters:          exporterMap,
		performanceMonitor: NewPerformanceMonitor(),
		workerPool:         workerPool,
		jobQueue:           make(chan WorkerJob, cfg.ProcessingOptions.Workers*2),
		resultQueue:        make(chan WorkerResult, cfg.ProcessingOptions.Workers*2),
	}
	
	log.Printf("MDB processor initialized with %d workers", cfg.ProcessingOptions.Workers)
	return processor, nil
}

// ProcessMDB processes an MDB file according to the request
func (p *Processor) ProcessMDB(req ProcessingRequest) (*ConversionJob, error) {
	log.Printf("Starting MDB processing: %s for %s", req.SourceFile, req.CompanyName)
	
	// Create conversion job
	job := &ConversionJob{
		ID:          generateJobID(),
		SourceFile:  req.SourceFile,
		CompanyName: req.CompanyName,
		StartTime:   time.Now(),
		Status:      "initializing",
		TableStats:  make(map[string]TableStats),
		ValidationStats: ValidationStats{
			ValidationsByType:  make(map[string]int),
			ValidationsByTable: make(map[string]int),
		},
	}
	
	p.setCurrentJob(job)
	
	// Start performance monitoring
	p.performanceMonitor.Start(job.ID)
	defer p.performanceMonitor.Stop()
	
	// Start worker pool
	p.workerPool.Start()
	defer p.workerPool.Stop()
	
	// Process in phases
	if err := p.processInPhases(req, job); err != nil {
		job.SetStatus("failed")
		job.AddError(ProcessingError{
			ErrorType:   "critical",
			Description: fmt.Sprintf("Processing failed: %v", err),
			Severity:    1,
		})
		return job, err
	}
	
	// Complete the job
	job.Complete()
	
	// Capture final performance stats
	job.PerformanceStats = p.performanceMonitor.GetStats()
	
	log.Printf("MDB processing completed: %s (Status: %s)", job.ID, job.GetStatus())
	return job, nil
}

// processInPhases executes the processing in distinct phases
func (p *Processor) processInPhases(req ProcessingRequest, job *ConversionJob) error {
	phases := []struct {
		name string
		fn   func(ProcessingRequest, *ConversionJob) error
	}{
		{"analyzing", p.phaseAnalyze},
		{"extracting", p.phaseExtract},
		{"validating", p.phaseValidate},
		{"transforming", p.phaseTransform},
		{"exporting", p.phaseExport},
	}
	
	for _, phase := range phases {
		log.Printf("Starting phase: %s", phase.name)
		job.SetStatus(phase.name)
		
		phaseStart := time.Now()
		if err := phase.fn(req, job); err != nil {
			return fmt.Errorf("phase %s failed: %w", phase.name, err)
		}
		
		phaseDuration := time.Since(phaseStart)
		log.Printf("Completed phase: %s (Duration: %v)", phase.name, phaseDuration)
		
		// Update performance stats
		p.performanceMonitor.RecordPhase(phase.name, phaseDuration)
		
		// Send progress update
		if req.ProgressChan != nil {
			req.ProgressChan <- ProgressUpdate{
				JobID:     job.ID,
				Phase:     phase.name,
				Message:   fmt.Sprintf("Completed %s phase", phase.name),
				Timestamp: time.Now(),
			}
		}
	}
	
	return nil
}

// phaseAnalyze analyzes the MDB file structure
func (p *Processor) phaseAnalyze(req ProcessingRequest, job *ConversionJob) error {
	log.Println("Analyzing MDB file structure...")
	
	// For this implementation, we'll work with CSV files extracted from MDB
	// In a real implementation, you'd use mdb-tools or similar
	
	analyzer := NewMDBAnalyzer(p.config)
	tables, err := analyzer.AnalyzeStructure(req.SourceFile)
	if err != nil {
		return fmt.Errorf("structure analysis failed: %w", err)
	}
	
	if len(tables) == 0 {
		return fmt.Errorf("no tables found in MDB file")
	}
	
	log.Printf("Found %d tables for processing", len(tables))
	
	// Store table information in job context
	p.setAnalysisResults(job, tables)
	
	return nil
}

// phaseExtract extracts data from the MDB file
func (p *Processor) phaseExtract(req ProcessingRequest, job *ConversionJob) error {
	log.Println("Extracting data from MDB file...")
	
	tables := p.getAnalysisResults(job)
	extractor := NewDataExtractor(p.config)
	
	for _, table := range tables {
		log.Printf("Extracting data from table: %s", table.Name)
		
		extractedData, err := extractor.ExtractTable(req.SourceFile, table)
		if err != nil {
			job.AddError(ProcessingError{
				Table:       table.Name,
				ErrorType:   "extraction",
				Description: fmt.Sprintf("Failed to extract table data: %v", err),
				Severity:    2,
			})
			continue
		}
		
		// Store extracted data for next phase
		p.setExtractionResults(job, table.Name, extractedData)
		
		// Update progress
		if req.ProgressChan != nil {
			req.ProgressChan <- ProgressUpdate{
				JobID:        job.ID,
				Table:        table.Name,
				Phase:        "extracting",
				Message:      fmt.Sprintf("Extracted %d records", len(extractedData)),
				Timestamp:    time.Now(),
			}
		}
	}
	
	return nil
}

// phaseValidate validates extracted data
func (p *Processor) phaseValidate(req ProcessingRequest, job *ConversionJob) error {
	log.Println("Validating extracted data...")
	
	tables := p.getAnalysisResults(job)
	
	for _, table := range tables {
		log.Printf("Validating table: %s", table.Name)
		
		extractedData := p.getExtractionResults(job, table.Name)
		if extractedData == nil {
			continue
		}
		
		validationResults, err := p.validator.ValidateTable(table, extractedData)
		if err != nil {
			job.AddError(ProcessingError{
				Table:       table.Name,
				ErrorType:   "validation",
				Description: fmt.Sprintf("Validation failed: %v", err),
				Severity:    2,
			})
			continue
		}
		
		// Process validation results
		p.processValidationResults(job, table.Name, validationResults, req.ErrorChan)
		
		// Update progress
		if req.ProgressChan != nil {
			req.ProgressChan <- ProgressUpdate{
				JobID:        job.ID,
				Table:        table.Name,
				Phase:        "validating",
				Message:      fmt.Sprintf("
