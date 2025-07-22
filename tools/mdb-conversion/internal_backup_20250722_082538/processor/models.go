// internal/processor/models.go
// Data structures for MDB processing operations
package processor

import (
	"sync"
	"time"
)

// ProcessingRequest represents a request to process an MDB file
type ProcessingRequest struct {
	SourceFile   string
	CompanyName  string
	OutputDir    string
	DryRun       bool
	ProgressChan chan<- ProgressUpdate
	ErrorChan    chan<- ProcessingError
}

// ConversionJob represents a complete conversion job
type ConversionJob struct {
	ID               string                 `json:"job_id"`
	SourceFile       string                 `json:"source_file"`
	CompanyName      string                 `json:"company_name"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Status           string                 `json:"status"`
	TablesProcessed  int                    `json:"tables_processed"`
	RecordsTotal     int                    `json:"records_total"`
	RecordsValid     int                    `json:"records_valid"`
	RecordsInvalid   int                    `json:"records_invalid"`
	OutputFiles      []string               `json:"output_files"`
	Errors           []ProcessingError      `json:"errors"`
	PerformanceStats *PerformanceStats      `json:"performance_stats,omitempty"`
	TableStats       map[string]TableStats  `json:"table_stats"`
	ValidationStats  ValidationStats        `json:"validation_stats"`
	mutex            sync.RWMutex           `json:"-"`
}

// ProgressUpdate represents a progress update during processing
type ProgressUpdate struct {
	JobID            string    `json:"job_id"`
	Table            string    `json:"table"`
	Phase            string    `json:"phase"` // "analyzing", "extracting", "validating", "writing"
	RecordsProcessed int       `json:"records_processed"`
	TotalRecords     int       `json:"total_records"`
	Percentage       float64   `json:"percentage"`
	Message          string    `json:"message"`
	Timestamp        time.Time `json:"timestamp"`
}

// ProcessingError represents an error that occurred during processing
type ProcessingError struct {
	JobID       string    `json:"job_id"`
	Table       string    `json:"table"`
	RecordID    string    `json:"record_id"`
	FieldName   string    `json:"field_name"`
	ErrorType   string    `json:"error_type"` // "validation", "transformation", "business_rule", "critical"
	Description string    `json:"description"`
	Value       string    `json:"value,omitempty"`
	Suggestion  string    `json:"suggestion,omitempty"`
	Severity    int       `json:"severity"` // 1=critical, 2=error, 3=warning
	Timestamp   time.Time `json:"timestamp"`
}

// TableInfo contains metadata about a source table
type TableInfo struct {
	Name          string             `json:"name"`
	OriginalName  string             `json:"original_name"`
	Columns       []ColumnInfo       `json:"columns"`
	RecordCount   int                `json:"record_count"`
	Relationships []RelationshipInfo `json:"relationships"`
	BusinessRules []string           `json:"business_rules"`
	Priority      int                `json:"priority"` // Processing order priority
	Dependencies  []string           `json:"dependencies"` // Tables this depends on
}

// ColumnInfo contains metadata about a column
type ColumnInfo struct {
	Name               string   `json:"name"`
	OriginalName       string   `json:"original_name"`
	OriginalType       string   `json:"original_type"`
	PostgreSQLType     string   `json:"postgresql_type"`
	IsNullable         bool     `json:"is_nullable"`
	HasDefault         bool     `json:"has_default"`
	DefaultValue       *string  `json:"default_value,omitempty"`
	BusinessRules      []string `json:"business_rules"`
	ValidationRules    []string `json:"validation_rules"`
	SampleValues       []string `json:"sample_values,omitempty"`
	UniqueValueCount   int      `json:"unique_value_count"`
	NullValueCount     int      `json:"null_value_count"`
	TransformationFunc string   `json:"transformation_func,omitempty"`
}

// RelationshipInfo represents a foreign key relationship
type RelationshipInfo struct {
	FromTable    string `json:"from_table"`
	FromColumn   string `json:"from_column"`
	ToTable      string `json:"to_table"`
	ToColumn     string `json:"to_column"`
	Constraint   string `json:"constraint"`
	Relationship string `json:"relationship"` // "one_to_one", "one_to_many", "many_to_many"
	Required     bool   `json:"required"`
}

// TableStats contains statistics about table processing
type TableStats struct {
	TableName        string        `json:"table_name"`
	RecordsTotal     int           `json:"records_total"`
	RecordsValid     int           `json:"records_valid"`
	RecordsInvalid   int           `json:"records_invalid"`
	ProcessingTime   time.Duration `json:"processing_time"`
	ValidationErrors int           `json:"validation_errors"`
	TransformErrors  int           `json:"transform_errors"`
	BusinessErrors   int           `json:"business_errors"`
	OutputFiles      []string      `json:"output_files"`
}

// ValidationStats contains overall validation statistics
type ValidationStats struct {
	TotalValidations     int                       `json:"total_validations"`
	PassedValidations    int                       `json:"passed_validations"`
	FailedValidations    int                       `json:"failed_validations"`
	ValidationsByType    map[string]int            `json:"validations_by_type"`
	ValidationsByTable   map[string]int            `json:"validations_by_table"`
	CriticalErrors       int                       `json:"critical_errors"`
	BusinessRuleViolations int                     `json:"business_rule_violations"`
	DataQualityScore     float64                   `json:"data_quality_score"`
	TopErrorTypes        []ErrorTypeCount          `json:"top_error_types"`
}

// ErrorTypeCount represents count of errors by type
type ErrorTypeCount struct {
	ErrorType string `json:"error_type"`
	Count     int    `json:"count"`
	Examples  []string `json:"examples,omitempty"`
}

// PerformanceStats contains performance metrics
type PerformanceStats struct {
	RecordsPerSecond   float64       `json:"records_per_second"`
	PeakMemoryUsage    int64         `json:"peak_memory_usage_bytes"`
	AvgMemoryUsage     int64         `json:"avg_memory_usage_bytes"`
	AvgCPUUsage        float64       `json:"avg_cpu_usage_percent"`
	DatabaseConnections int          `json:"database_connections"`
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	PhaseTimings       map[string]time.Duration `json:"phase_timings"`
	ThroughputByTable  map[string]float64 `json:"throughput_by_table"`
}

// ConversionResult represents the result of a table conversion
type ConversionResult struct {
	TableName    string        `json:"table_name"`
	Success      bool          `json:"success"`
	RecordsTotal int           `json:"records_total"`
	RecordsValid int           `json:"records_valid"`
	OutputFiles  []string      `json:"output_files"`
	Errors       []ProcessingError `json:"errors"`
	Duration     time.Duration `json:"duration"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DatabaseConnection represents a database connection configuration
type DatabaseConnection struct {
	ConnectionString string
	MaxConnections   int
	IdleConnections  int
	Schema          string
}

// WorkerJob represents a job for a worker goroutine
type WorkerJob struct {
	Type     string      `json:"type"` // "extract", "transform", "validate", "load"
	Table    TableInfo   `json:"table"`
	Data     interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

// WorkerResult represents the result from a worker
type WorkerResult struct {
	Job     WorkerJob         `json:"job"`
	Success bool              `json:"success"`
	Result  ConversionResult  `json:"result"`
	Error   error             `json:"error,omitempty"`
}

// Methods for ConversionJob

// HasErrors returns true if the job has any errors
func (j *ConversionJob) HasErrors() bool {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	return len(j.Errors) > 0
}

// HasCriticalErrors returns true if the job has critical errors
func (j *ConversionJob) HasCriticalErrors() bool {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	
	for _, err := range j.Errors {
		if err.Severity == 1 || err.ErrorType == "critical" {
			return true
		}
	}
	return false
}

// SuccessRate returns the success rate as a percentage
func (j *ConversionJob) SuccessRate() float64 {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	
	if j.RecordsTotal == 0 {
		return 0.0
	}
	return float64(j.RecordsValid) / float64(j.RecordsTotal) * 100.0
}

// AddError adds an error to the job
func (j *ConversionJob) AddError(err ProcessingError) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	
	err.JobID = j.ID
	err.Timestamp = time.Now()
	j.Errors = append(j.Errors, err)
	
	// Update invalid record count
	if err.Severity <= 2 {
		j.RecordsInvalid++
	}
}

// AddOutputFile adds an output file to the job
func (j *ConversionJob) AddOutputFile(filepath string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.OutputFiles = append(j.OutputFiles, filepath)
}

// UpdateTableStats updates statistics for a specific table
func (j *ConversionJob) UpdateTableStats(tableName string, stats TableStats) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	
	if j.TableStats == nil {
		j.TableStats = make(map[string]TableStats)
	}
	j.TableStats[tableName] = stats
	
	// Update overall totals
	j.RecordsTotal += stats.RecordsTotal
	j.RecordsValid += stats.RecordsValid
	j.RecordsInvalid += stats.RecordsInvalid
}

// SetStatus sets the job status thread-safely
func (j *ConversionJob) SetStatus(status string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.Status = status
}

// GetStatus gets the job status thread-safely
func (j *ConversionJob) GetStatus() string {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	return j.Status
}

// Complete marks the job as complete
func (j *ConversionJob) Complete() {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	
	now := time.Now()
	j.EndTime = &now
	
	if j.HasCriticalErrors() {
		j.Status = "failed"
	} else if j.HasErrors() {
		j.Status = "completed_with_warnings"
	} else {
		j.Status = "completed"
	}
}

// Methods for TableInfo

// GetColumnByName returns a column by name
func (t *TableInfo) GetColumnByName(name string) (ColumnInfo, bool) {
	for _, col := range t.Columns {
		if col.Name == name || col.OriginalName == name {
			return col, true
		}
	}
	return ColumnInfo{}, false
}

// HasDependencies returns true if the table has dependencies
func (t *TableInfo) HasDependencies() bool {
	return len(t.Dependencies) > 0
}

// GetPrimaryKeyColumns returns columns that are primary keys
func (t *TableInfo) GetPrimaryKeyColumns() []ColumnInfo {
	var pkColumns []ColumnInfo
	for _, col := range t.Columns {
		// Check if column name suggests it's a primary key
		if col.Name == "id" || col.Name == t.Name+"_id" {
			pkColumns = append(pkColumns, col)
		}
	}
	return pkColumns
}

// GetForeignKeyColumns returns columns that are foreign keys
func (t *TableInfo) GetForeignKeyColumns() []ColumnInfo {
	var fkColumns []ColumnInfo
	for _, col := range t.Columns {
		// Check if column name suggests it's a foreign key
		if len(col.Name) > 3 && col.Name[len(col.Name)-3:] == "_id" && col.Name != t.Name+"_id" {
			fkColumns = append(fkColumns, col)
		}
	}
	return fkColumns
}

// Methods for ProcessingError

// IsCritical returns true if the error is critical
func (e *ProcessingError) IsCritical() bool {
	return e.Severity == 1 || e.ErrorType == "critical"
}

// IsWarning returns true if the error is a warning
func (e *ProcessingError) IsWarning() bool {
	return e.Severity == 3
}

// String returns a string representation of the error
func (e *ProcessingError) String() string {
	return fmt.Sprintf("[%s] %s.%s: %s", e.ErrorType, e.Table, e.FieldName, e.Description)
}

// Methods for ValidationStats

// CalculateQualityScore calculates the data quality score
func (v *ValidationStats) CalculateQualityScore() {
	if v.TotalValidations == 0 {
		v.DataQualityScore = 0.0
		return
	}
	
	// Base score from passed validations
	baseScore := float64(v.PassedValidations) / float64(v.TotalValidations) * 100.0
	
	// Penalty for critical errors
	criticalPenalty := float64(v.CriticalErrors) * 5.0
	
	// Penalty for business rule violations
	businessPenalty := float64(v.BusinessRuleViolations) * 2.0
	
	// Calculate final score
	v.DataQualityScore = baseScore - criticalPenalty - businessPenalty
	
	// Ensure score is between 0 and 100
	if v.DataQualityScore < 0 {
		v.DataQualityScore = 0.0
	}
	if v.DataQualityScore > 100 {
		v.DataQualityScore = 100.0
	}
}

// AddValidationResult adds a validation result to the stats
func (v *ValidationStats) AddValidationResult(errorType string, table string, passed bool) {
	v.TotalValidations++
	
	if passed {
		v.PassedValidations++
	} else {
		v.FailedValidations++
	}
	
	// Initialize maps if needed
	if v.ValidationsByType == nil {
		v.ValidationsByType = make(map[string]int)
	}
	if v.ValidationsByTable == nil {
		v.ValidationsByTable = make(map[string]int)
	}
	
	// Update counts
	if !passed {
		v.ValidationsByType[errorType]++
		v.ValidationsByTable[table]++
		
		if errorType == "critical" {
			v.CriticalErrors++
		}
		if errorType == "business_rule" {
			v.BusinessRuleViolations++
		}
	}
	
	// Recalculate quality score
	v.CalculateQualityScore()
}
