package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Extension represents a DuckDB extension with its metadata and status
type Extension struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Loaded      bool   `json:"loaded"`
	Installed   bool   `json:"installed"`
	BuiltIn     bool   `json:"built_in,omitempty"`
	Version     string `json:"version,omitempty"`
}

// ExtensionConfig holds configuration for extension management
type ExtensionConfig struct {
	// AutoInstall automatically installs extensions when loading
	AutoInstall bool

	// PreloadExtensions list of extensions to load on database connection
	PreloadExtensions []string

	// Timeout for extension operations (0 = no timeout)
	Timeout time.Duration

	// RepositoryURL custom extension repository URL
	RepositoryURL string

	// AllowUnsigned allows loading unsigned extensions (security risk)
	AllowUnsigned bool
}

// ExtensionManager handles DuckDB extension operations
type ExtensionManager struct {
	db     *gorm.DB
	config *ExtensionConfig
}

// Common DuckDB extensions
const (
	// Core Extensions (built-in)
	ExtensionJSON    = "json"
	ExtensionParquet = "parquet"
	ExtensionICU     = "icu"

	// Analytics Extensions
	ExtensionAutoComplete = "autocomplete"
	ExtensionFTS          = "fts"
	ExtensionTPCH         = "tpch"
	ExtensionTPCDS        = "tpcds"

	// Data Format Extensions
	ExtensionCSV    = "csv"
	ExtensionExcel  = "excel"
	ExtensionArrow  = "arrow"
	ExtensionSQLite = "sqlite"

	// Networking Extensions
	ExtensionHTTPS = "httpfs"
	ExtensionS3    = "aws"
	ExtensionAzure = "azure"

	// Geospatial Extensions
	ExtensionSpatial = "spatial"

	// Machine Learning Extensions
	ExtensionML = "ml"

	// Time Series Extensions
	ExtensionTimeSeries = "timeseries"

	// Visualization Extensions
	ExtensionVisualization = "visualization"
)

// NewExtensionManager creates a new extension manager instance
func NewExtensionManager(db *gorm.DB, config *ExtensionConfig) *ExtensionManager {
	if config == nil {
		config = &ExtensionConfig{
			AutoInstall: true,
			Timeout:     30 * time.Second,
		}
	}

	return &ExtensionManager{
		db:     db,
		config: config,
	}
}

// ListExtensions returns all available extensions
func (m *ExtensionManager) ListExtensions() ([]Extension, error) {
	ctx := context.Background()
	if m.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.config.Timeout)
		defer cancel()
	}

	var extensions []Extension

	// Query duckdb_extensions() function to get extension information
	query := `
		SELECT 
			extension_name as name,
			loaded,
			installed,
			description
		FROM duckdb_extensions()
		ORDER BY extension_name
	`

	rows, err := m.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query extensions: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var ext Extension
		var description sql.NullString

		if err := rows.Scan(&ext.Name, &ext.Loaded, &ext.Installed, &description); err != nil {
			return nil, fmt.Errorf("failed to scan extension row: %w", err)
		}

		if description.Valid {
			ext.Description = description.String
		}

		extensions = append(extensions, ext)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating extension rows: %w", err)
	}

	return extensions, nil
}

// GetExtension returns information about a specific extension
func (m *ExtensionManager) GetExtension(name string) (*Extension, error) {
	ctx := context.Background()
	if m.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.config.Timeout)
		defer cancel()
	}

	query := `
		SELECT 
			extension_name as name,
			loaded,
			installed,
			description
		FROM duckdb_extensions()
		WHERE extension_name = ?
	`

	var ext Extension
	var description sql.NullString

	err := m.db.WithContext(ctx).Raw(query, name).Row().Scan(
		&ext.Name, &ext.Loaded, &ext.Installed, &description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("extension '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get extension '%s': %w", name, err)
	}

	if description.Valid {
		ext.Description = description.String
	}

	return &ext, nil
}

// LoadExtension loads an extension, optionally installing it first
func (m *ExtensionManager) LoadExtension(name string) error {
	ctx := context.Background()
	if m.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.config.Timeout)
		defer cancel()
	}

	// Check if extension is already loaded
	if m.IsExtensionLoaded(name) {
		return nil // Already loaded
	}

	// Install extension if auto-install is enabled and extension is not installed
	if m.config.AutoInstall {
		ext, err := m.GetExtension(name)
		if err != nil {
			return fmt.Errorf("failed to check extension status: %w", err)
		}

		if !ext.Installed {
			if err := m.InstallExtension(name); err != nil {
				return fmt.Errorf("failed to install extension '%s': %w", name, err)
			}
		}
	}

	// Load the extension
	query := fmt.Sprintf("LOAD %s", m.quoteName(name))
	if err := m.db.WithContext(ctx).Exec(query).Error; err != nil {
		return fmt.Errorf("failed to load extension '%s': %w", name, err)
	}

	return nil
}

// InstallExtension installs an extension from the repository
func (m *ExtensionManager) InstallExtension(name string) error {
	ctx := context.Background()
	if m.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.config.Timeout)
		defer cancel()
	}

	// Check if already installed
	ext, err := m.GetExtension(name)
	if err == nil && ext.Installed {
		return nil // Already installed
	}

	// Install the extension
	query := fmt.Sprintf("INSTALL %s", m.quoteName(name))
	if err := m.db.WithContext(ctx).Exec(query).Error; err != nil {
		return fmt.Errorf("failed to install extension '%s': %w", name, err)
	}

	return nil
}

// IsExtensionLoaded checks if an extension is currently loaded
func (m *ExtensionManager) IsExtensionLoaded(name string) bool {
	ext, err := m.GetExtension(name)
	if err != nil {
		return false
	}
	return ext.Loaded
}

// GetLoadedExtensions returns all currently loaded extensions
func (m *ExtensionManager) GetLoadedExtensions() ([]Extension, error) {
	extensions, err := m.ListExtensions()
	if err != nil {
		return nil, err
	}

	var loaded []Extension
	for _, ext := range extensions {
		if ext.Loaded {
			loaded = append(loaded, ext)
		}
	}

	return loaded, nil
}

// LoadExtensions loads multiple extensions
func (m *ExtensionManager) LoadExtensions(names []string) error {
	for _, name := range names {
		if err := m.LoadExtension(name); err != nil {
			return fmt.Errorf("failed to load extension '%s': %w", name, err)
		}
	}
	return nil
}

// PreloadExtensions loads all configured preload extensions
func (m *ExtensionManager) PreloadExtensions() error {
	if len(m.config.PreloadExtensions) == 0 {
		return nil
	}

	return m.LoadExtensions(m.config.PreloadExtensions)
}

// quoteName safely quotes an extension name for SQL
func (m *ExtensionManager) quoteName(name string) string {
	// Remove any potentially dangerous characters
	cleaned := strings.ReplaceAll(name, "'", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")
	cleaned = strings.ReplaceAll(cleaned, ";", "")
	cleaned = strings.ReplaceAll(cleaned, "--", "")
	return cleaned
}

// ExtensionHelper provides convenience methods for common extension operations
type ExtensionHelper struct {
	manager *ExtensionManager
}

// NewExtensionHelper creates a new extension helper
func NewExtensionHelper(manager *ExtensionManager) *ExtensionHelper {
	return &ExtensionHelper{manager: manager}
}

// EnableAnalytics loads common analytics extensions
func (h *ExtensionHelper) EnableAnalytics() error {
	analyticsExtensions := []string{
		ExtensionJSON,
		ExtensionParquet,
		ExtensionFTS,
		ExtensionAutoComplete,
	}

	return h.manager.LoadExtensions(analyticsExtensions)
}

// EnableDataFormats loads common data format extensions
func (h *ExtensionHelper) EnableDataFormats() error {
	// Only try to load extensions that are commonly available
	formatExtensions := []string{
		ExtensionJSON,
		ExtensionParquet,
	}

	// Try to load other extensions, but don't fail if they're not available
	optionalExtensions := []string{
		ExtensionCSV,
		ExtensionExcel,
		ExtensionArrow,
	}

	// Load essential extensions first
	if err := h.manager.LoadExtensions(formatExtensions); err != nil {
		return err
	}

	// Try optional extensions, log failures but don't return error
	for _, ext := range optionalExtensions {
		if err := h.manager.LoadExtension(ext); err != nil {
			// Log but don't fail - these might not be available in all builds
			continue
		}
	}

	return nil
}

// EnableCloudAccess loads cloud storage extensions
func (h *ExtensionHelper) EnableCloudAccess() error {
	cloudExtensions := []string{
		ExtensionHTTPS,
		ExtensionS3,
		ExtensionAzure,
	}

	return h.manager.LoadExtensions(cloudExtensions)
}

// EnableSpatial loads geospatial extensions
func (h *ExtensionHelper) EnableSpatial() error {
	return h.manager.LoadExtension(ExtensionSpatial)
}

// EnableMachineLearning loads ML extensions
func (h *ExtensionHelper) EnableMachineLearning() error {
	return h.manager.LoadExtension(ExtensionML)
}

// EnableTimeSeries loads time series extensions
func (h *ExtensionHelper) EnableTimeSeries() error {
	return h.manager.LoadExtension(ExtensionTimeSeries)
}

// Dialector integration for extensions

// extensionAwareDialector wraps the standard dialector with extension support
type extensionAwareDialector struct {
	*Dialector
	extensionConfig *ExtensionConfig
	manager         *ExtensionManager
}

// NewWithExtensions creates a new dialector with extension support
func NewWithExtensions(config Config, extensionConfig *ExtensionConfig) gorm.Dialector {
	return &extensionAwareDialector{
		Dialector:       &Dialector{Config: &config},
		extensionConfig: extensionConfig,
	}
}

// OpenWithExtensions creates a dialector with extension support using DSN
func OpenWithExtensions(dsn string, extensionConfig *ExtensionConfig) gorm.Dialector {
	return NewWithExtensions(Config{DSN: dsn}, extensionConfig)
}

// Initialize initializes the dialector with extension support
func (d *extensionAwareDialector) Initialize(db *gorm.DB) error {
	// First initialize the base dialector
	if err := d.Dialector.Initialize(db); err != nil {
		return err
	}

	// Create extension manager but don't preload yet
	if d.extensionConfig != nil {
		d.manager = NewExtensionManager(db, d.extensionConfig)

		// We'll preload extensions lazily when first accessed
		// or provide a separate method to trigger preloading
	}

	return nil
}

// Extension manager retrieval functions

// GetExtensionManager retrieves the extension manager from a database instance
func GetExtensionManager(db *gorm.DB) (*ExtensionManager, error) {
	// Use type assertion on the dialector to get the extension manager
	if extensionDialector, ok := db.Dialector.(*extensionAwareDialector); ok {
		if extensionDialector.manager != nil {
			return extensionDialector.manager, nil
		}
		return nil, fmt.Errorf("extension manager not initialized - check extension configuration")
	}
	return nil, fmt.Errorf("extension manager not found - use NewWithExtensions or OpenWithExtensions")
}

// InitializeExtensions manually triggers preloading of configured extensions
// This should be called after the database connection is fully established
func InitializeExtensions(db *gorm.DB) error {
	manager, err := GetExtensionManager(db)
	if err != nil {
		return err
	}
	return manager.PreloadExtensions()
}

// MustGetExtensionManager retrieves the extension manager, panics if not found
func MustGetExtensionManager(db *gorm.DB) *ExtensionManager {
	manager, err := GetExtensionManager(db)
	if err != nil {
		panic(err)
	}
	return manager
}
