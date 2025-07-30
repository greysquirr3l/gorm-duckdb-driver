package duckdb_test

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// setupExtensionsTestDB creates a test database connection following GORM best practices
func setupExtensionsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

// cleanupExtensionsTestDB closes the database connection properly
func cleanupExtensionsTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

func TestExtensionManager_BasicOperations(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	// Create extension manager with default config
	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
		Timeout:     0, // Use default
	}
	manager := duckdb.NewExtensionManager(db, config)

	// Test listing extensions
	extensions, err := manager.ListExtensions()
	if err != nil {
		t.Fatalf("Failed to list extensions: %v", err)
	}

	if len(extensions) == 0 {
		t.Error("Expected at least some extensions to be available")
	}

	// Verify we have some built-in extensions
	var foundJSON, foundParquet bool
	for _, ext := range extensions {
		if ext.Name == duckdb.ExtensionJSON {
			foundJSON = true
		}
		if ext.Name == duckdb.ExtensionParquet {
			foundParquet = true
		}
	}

	if !foundJSON {
		t.Error("Expected JSON extension to be available")
	}
	if !foundParquet {
		t.Error("Expected Parquet extension to be available")
	}
}

func TestExtensionManager_LoadExtension(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)

	// Test loading JSON extension (should be built-in)
	err := manager.LoadExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Failed to load JSON extension: %v", err)
	}

	// Verify it's loaded
	ext, err := manager.GetExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Failed to get JSON extension info: %v", err)
	}

	if !ext.Loaded {
		t.Error("JSON extension should be loaded")
	}

	// Test that loading again doesn't fail (idempotent)
	err = manager.LoadExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Loading already loaded extension should not fail: %v", err)
	}
}

func TestExtensionManager_GetExtension(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	manager := duckdb.NewExtensionManager(db, nil)

	// Test getting existing extension
	ext, err := manager.GetExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Failed to get JSON extension: %v", err)
	}

	if ext.Name != duckdb.ExtensionJSON {
		t.Errorf("Expected extension name %s, got %s", duckdb.ExtensionJSON, ext.Name)
	}

	// Test getting non-existent extension
	_, err = manager.GetExtension("nonexistent_extension")
	if err == nil {
		t.Error("Expected error when getting non-existent extension")
	}
}

func TestExtensionManager_GetLoadedExtensions(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)

	// Load some extensions
	loadTestExtensions(t, manager)

	// Get loaded extensions
	loaded, err := manager.GetLoadedExtensions()
	if err != nil {
		t.Fatalf("Failed to get loaded extensions: %v", err)
	}

	// Verify loaded extensions
	validateLoadedExtensions(t, loaded)
}

func loadTestExtensions(t *testing.T, manager *duckdb.ExtensionManager) {
	if err := manager.LoadExtension(duckdb.ExtensionJSON); err != nil {
		t.Fatalf("Failed to load JSON extension: %v", err)
	}

	if err := manager.LoadExtension(duckdb.ExtensionParquet); err != nil {
		t.Fatalf("Failed to load Parquet extension: %v", err)
	}
}

func validateLoadedExtensions(t *testing.T, loaded []duckdb.Extension) {
	// Should have at least the ones we loaded
	foundJSON := findLoadedExtension(loaded, duckdb.ExtensionJSON)
	foundParquet := findLoadedExtension(loaded, duckdb.ExtensionParquet)

	if !foundJSON {
		t.Error("JSON extension should be in loaded extensions list")
	}
	if !foundParquet {
		t.Error("Parquet extension should be in loaded extensions list")
	}
}

func findLoadedExtension(extensions []duckdb.Extension, name string) bool {
	for _, ext := range extensions {
		if ext.Name == name && ext.Loaded {
			return true
		}
	}
	return false
}

func TestExtensionManager_IsExtensionLoaded(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)

	// Initially should not be loaded (or might be auto-loaded)
	initiallyLoaded := manager.IsExtensionLoaded(duckdb.ExtensionJSON)

	// Load the extension
	err := manager.LoadExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Failed to load JSON extension: %v", err)
	}

	// Now should definitely be loaded
	if !manager.IsExtensionLoaded(duckdb.ExtensionJSON) {
		t.Error("JSON extension should be loaded after LoadExtension call")
	}

	if !initiallyLoaded {
		t.Log("JSON extension was not initially loaded, then successfully loaded")
	} else {
		t.Log("JSON extension was already loaded (auto-loaded)")
	}
}

func TestExtensionHelper_EnableAnalytics(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)
	helper := duckdb.NewExtensionHelper(manager)

	// Enable analytics extensions
	err := helper.EnableAnalytics()
	if err != nil {
		t.Fatalf("Failed to enable analytics extensions: %v", err)
	}

	// Verify at least some core analytics extensions are loaded
	essentialExtensions := []string{duckdb.ExtensionJSON, duckdb.ExtensionParquet}
	for _, extName := range essentialExtensions {
		if !manager.IsExtensionLoaded(extName) {
			t.Errorf("Essential analytics extension %s should be loaded", extName)
		}
	}
}

func TestExtensionHelper_EnableDataFormats(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)
	helper := duckdb.NewExtensionHelper(manager)

	// Enable data format extensions
	err := helper.EnableDataFormats()
	if err != nil {
		t.Fatalf("Failed to enable data format extensions: %v", err)
	}

	// Verify core format extensions are loaded
	formatExtensions := []string{duckdb.ExtensionJSON, duckdb.ExtensionParquet}
	for _, extName := range formatExtensions {
		if !manager.IsExtensionLoaded(extName) {
			t.Errorf("Data format extension %s should be loaded", extName)
		}
	}
}

func TestExtensionHelper_EnableSpatial(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)
	helper := duckdb.NewExtensionHelper(manager)

	// Try to enable spatial extension
	err := helper.EnableSpatial()
	if err != nil {
		// Spatial extension might not be available in all builds
		t.Logf("Could not enable spatial extension (may not be available): %v", err)
		return
	}

	// If successful, verify it's loaded
	if !manager.IsExtensionLoaded(duckdb.ExtensionSpatial) {
		t.Error("Spatial extension should be loaded after EnableSpatial")
	}
}

// TODO: Fix dialector integration tests - currently having InstanceSet timing issues
/*
func TestDialectorWithExtensions(t *testing.T) {
	// Test creating dialector with extension support
	extensionConfig := &duckdb.ExtensionConfig{
		AutoInstall:       true,
		PreloadExtensions: []string{duckdb.ExtensionJSON, duckdb.ExtensionParquet},
	}

	dialector := NewWithExtensions(Config{
		DSN: ":memory:",
	}, extensionConfig)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database with extensions: %v", err)
	}
	defer cleanupExtensionsTestDB(db)

	// Verify extension manager is available
	manager, err := duckdb.GetExtensionManager(db)
	if err != nil {
		t.Fatalf("Failed to get extension manager: %v", err)
	}

	// Verify preloaded extensions are loaded
	if !manager.IsExtensionLoaded(duckdb.ExtensionJSON) {
		t.Error("JSON extension should be preloaded")
	}
	if !manager.IsExtensionLoaded(duckdb.ExtensionParquet) {
		t.Error("Parquet extension should be preloaded")
	}
}

func TestOpenWithExtensions(t *testing.T) {
	extensionConfig := &duckdb.ExtensionConfig{
		AutoInstall:       true,
		PreloadExtensions: []string{duckdb.ExtensionJSON},
	}

	dialector := OpenWithExtensions(":memory:", extensionConfig)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database with extensions: %v", err)
	}
	defer cleanupExtensionsTestDB(db)

	// Verify extension manager is available
	manager, err := duckdb.GetExtensionManager(db)
	if err != nil {
		t.Fatalf("Failed to get extension manager: %v", err)
	}

	// Verify preloaded extension is loaded
	if !manager.IsExtensionLoaded(duckdb.ExtensionJSON) {
		t.Error("JSON extension should be preloaded")
	}
}
*/

func TestExtensionWithoutConfig(t *testing.T) {
	// Test that normal dialector still works without extension config
	dialector := duckdb.Open(":memory:")

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database without extensions: %v", err)
	}
	defer cleanupExtensionsTestDB(db)

	// Extension manager should not be available
	_, err = duckdb.GetExtensionManager(db)
	if err == nil {
		t.Error("Expected error when getting extension manager without config")
	}
}

func TestMustGetExtensionManager_Panic(t *testing.T) {
	// Test that MustGetExtensionManager panics when extension manager is not available
	dialector := duckdb.Open(":memory:")

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer cleanupExtensionsTestDB(db)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustGetExtensionManager to panic")
		}
	}()

	duckdb.MustGetExtensionManager(db)
}

func TestExtensionFunctionalUsage(t *testing.T) {
	// Test that extensions actually work for real functionality
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall:       true,
		PreloadExtensions: []string{duckdb.ExtensionJSON},
	}
	manager := duckdb.NewExtensionManager(db, config)

	// Load the JSON extension manually
	err := manager.LoadExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Failed to load JSON extension: %v", err)
	}

	// Test JSON functionality (requires JSON extension)
	var result string
	err = db.Raw("SELECT json_type('null') as json_result").Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to use JSON function: %v", err)
	}

	if result != "NULL" {
		t.Errorf("Expected 'NULL', got '%s'", result)
	}

	t.Logf("JSON function result: %s", result)
}

func TestExtensionManager_LoadMultipleExtensions(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
	}
	manager := duckdb.NewExtensionManager(db, config)

	// Load multiple extensions at once
	extensions := []string{duckdb.ExtensionJSON, duckdb.ExtensionParquet}
	err := manager.LoadExtensions(extensions)
	if err != nil {
		t.Fatalf("Failed to load multiple extensions: %v", err)
	}

	// Verify all extensions are loaded
	for _, extName := range extensions {
		if !manager.IsExtensionLoaded(extName) {
			t.Errorf("Extension %s should be loaded", extName)
		}
	}
}

func TestExtensionConfig_Defaults(t *testing.T) {
	db := setupExtensionsTestDB(t)
	defer cleanupExtensionsTestDB(db)

	// Test with nil config (should use defaults)
	manager := duckdb.NewExtensionManager(db, nil)

	// Test that the manager works with default config by trying to load an extension
	err := manager.LoadExtension(duckdb.ExtensionJSON)
	if err != nil {
		t.Fatalf("Failed to load extension with default config: %v", err)
	}

	// Verify the extension was loaded (this indirectly tests that AutoInstall defaults work)
	if !manager.IsExtensionLoaded(duckdb.ExtensionJSON) {
		t.Error("Extension should be loaded with default AutoInstall behavior")
	}
}
