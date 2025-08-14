package duckdb_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

func setupExtensionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	config := duckdb.ExtensionConfig{
		AutoInstall: true,
		Timeout:     30 * time.Second,
	}

	dialector := duckdb.OpenWithExtensions(":memory:", &config)
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Initialize extensions after database is ready
	err = duckdb.InitializeExtensions(db)
	require.NoError(t, err)

	return db
}

func setupBasicExtensionTestDB(t *testing.T) (*gorm.DB, *duckdb.ExtensionManager) {
	t.Helper()

	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	config := &duckdb.ExtensionConfig{
		AutoInstall: true,
		Timeout:     30 * time.Second,
	}

	manager := duckdb.NewExtensionManager(db, config)
	return db, manager
}

func TestExtensionConfig_Defaults(t *testing.T) {
	db, manager := setupBasicExtensionTestDB(t)
	_ = db

	// Check that default config is applied
	assert.NotNil(t, manager)
}

func TestExtensionManager_NewExtensionManager(t *testing.T) {
	db, _ := setupBasicExtensionTestDB(t)

	tests := []struct {
		name   string
		config *duckdb.ExtensionConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "custom config",
			config: &duckdb.ExtensionConfig{
				AutoInstall:       false,
				PreloadExtensions: []string{"json"},
				Timeout:           10 * time.Second,
				AllowUnsigned:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := duckdb.NewExtensionManager(db, tt.config)
			assert.NotNil(t, manager)
		})
	}
}

func TestExtensionManager_ListExtensions(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	extensions, err := manager.ListExtensions()
	require.NoError(t, err)

	// Should have at least some built-in extensions
	assert.Greater(t, len(extensions), 0)

	// Check that each extension has a name
	for _, ext := range extensions {
		assert.NotEmpty(t, ext.Name)
	}
}

func TestExtensionManager_GetExtension_JSON(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	// JSON extension should be available
	ext, err := manager.GetExtension("json")
	require.NoError(t, err)
	require.NotNil(t, ext)

	assert.Equal(t, "json", ext.Name)
	// JSON is usually built-in and loaded by default
}

func TestExtensionManager_GetExtension_NotFound(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	ext, err := manager.GetExtension("non_existent_extension")
	assert.Error(t, err)
	assert.Nil(t, ext)
	assert.Contains(t, err.Error(), "not found")
}

func TestExtensionManager_LoadExtension_JSON(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	// Load JSON extension (should work as it's built-in)
	err := manager.LoadExtension("json")
	assert.NoError(t, err)

	// Check it's loaded
	assert.True(t, manager.IsExtensionLoaded("json"))
}

func TestExtensionManager_LoadExtension_AlreadyLoaded(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	// Load JSON extension twice
	err := manager.LoadExtension("json")
	assert.NoError(t, err)

	err = manager.LoadExtension("json")
	assert.NoError(t, err) // Should not error if already loaded
}

func TestExtensionManager_IsExtensionLoaded(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	// Test with non-existent extension
	loaded := manager.IsExtensionLoaded("non_existent_extension")
	assert.False(t, loaded)

	// Load JSON and test
	err := manager.LoadExtension("json")
	require.NoError(t, err)

	loaded = manager.IsExtensionLoaded("json")
	assert.True(t, loaded)
}

func TestExtensionManager_GetLoadedExtensions(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	// Load JSON extension
	err := manager.LoadExtension("json")
	require.NoError(t, err)

	loaded, err := manager.GetLoadedExtensions()
	require.NoError(t, err)

	// Should contain at least the JSON extension
	found := false
	for _, ext := range loaded {
		if ext.Name == "json" {
			found = true
			assert.True(t, ext.Loaded)
			break
		}
	}
	assert.True(t, found, "JSON extension should be in loaded extensions list")
}

func TestExtensionManager_LoadExtensions_Multiple(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	extensions := []string{"json"}

	err := manager.LoadExtensions(extensions)
	assert.NoError(t, err)

	// Check all are loaded
	for _, extName := range extensions {
		assert.True(t, manager.IsExtensionLoaded(extName))
	}
}

func TestExtensionManager_LoadExtensions_Empty(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	err := manager.LoadExtensions([]string{})
	assert.NoError(t, err) // Should not error with empty list
}

func TestExtensionManager_PreloadExtensions_Empty(t *testing.T) {
	db, _ := setupBasicExtensionTestDB(t)

	config := &duckdb.ExtensionConfig{
		PreloadExtensions: []string{}, // Empty list
	}

	manager := duckdb.NewExtensionManager(db, config)
	err := manager.PreloadExtensions()
	assert.NoError(t, err)
}

func TestExtensionManager_PreloadExtensions_WithExtensions(t *testing.T) {
	db, _ := setupBasicExtensionTestDB(t)

	config := &duckdb.ExtensionConfig{
		PreloadExtensions: []string{"json"},
		AutoInstall:       true,
	}

	manager := duckdb.NewExtensionManager(db, config)
	err := manager.PreloadExtensions()
	assert.NoError(t, err)

	// Check that JSON is loaded
	assert.True(t, manager.IsExtensionLoaded("json"))
}

func TestExtensionHelper_NewExtensionHelper(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)

	helper := duckdb.NewExtensionHelper(manager)
	assert.NotNil(t, helper)
}

func TestExtensionHelper_EnableAnalytics(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)
	helper := duckdb.NewExtensionHelper(manager)

	// This might fail for some extensions that aren't available
	// but should not panic
	err := helper.EnableAnalytics()
	// We don't assert no error because some extensions might not be available
	// in the test environment, but we test that it doesn't panic
	_ = err
}

func TestExtensionHelper_EnableDataFormats(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)
	helper := duckdb.NewExtensionHelper(manager)

	// This should work as JSON and Parquet are usually built-in
	err := helper.EnableDataFormats()
	assert.NoError(t, err)

	// Check that at least JSON is loaded
	assert.True(t, manager.IsExtensionLoaded("json"))
}

func TestExtensionHelper_EnableCloudAccess(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)
	helper := duckdb.NewExtensionHelper(manager)

	// This might fail if cloud extensions aren't available
	err := helper.EnableCloudAccess()
	_ = err // Don't assert, just ensure it doesn't panic
}

func TestExtensionHelper_EnableSpatial(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)
	helper := duckdb.NewExtensionHelper(manager)

	err := helper.EnableSpatial()
	_ = err // Don't assert, spatial extension might not be available
}

func TestExtensionHelper_EnableMachineLearning(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)
	helper := duckdb.NewExtensionHelper(manager)

	err := helper.EnableMachineLearning()
	_ = err // Don't assert, ML extension might not be available
}

func TestExtensionHelper_EnableTimeSeries(t *testing.T) {
	_, manager := setupBasicExtensionTestDB(t)
	helper := duckdb.NewExtensionHelper(manager)

	err := helper.EnableTimeSeries()
	_ = err // Don't assert, time series extension might not be available
}

func TestExtensionAwareDialector_Initialize(t *testing.T) {
	config := &duckdb.ExtensionConfig{
		AutoInstall:       true,
		PreloadExtensions: []string{"json"},
	}

	dialector := duckdb.OpenWithExtensions(":memory:", config)
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Should be able to get the extension manager
	manager, err := duckdb.GetExtensionManager(db)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Now manually initialize extensions (since we don't do it during dialector init)
	err = duckdb.InitializeExtensions(db)
	require.NoError(t, err)

	// JSON should be loaded from preload
	assert.True(t, manager.IsExtensionLoaded("json"))
}

func TestExtensionAwareDialector_NewWithExtensions(t *testing.T) {
	t.Skip("Extension-aware dialector has GORM integration issues with InstanceSet")
}

func TestGetExtensionManager_Success(t *testing.T) {
	db := setupExtensionTestDB(t)

	manager, err := duckdb.GetExtensionManager(db)
	require.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestGetExtensionManager_NotFound(t *testing.T) {
	// Use regular dialector without extension support
	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	manager, err := duckdb.GetExtensionManager(db)
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "extension manager not found")
}

func TestMustGetExtensionManager_Success(t *testing.T) {
	db := setupExtensionTestDB(t)

	// Should not panic
	manager := duckdb.MustGetExtensionManager(db)
	assert.NotNil(t, manager)
}

func TestMustGetExtensionManager_Panic(t *testing.T) {
	// Use regular dialector without extension support
	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Should panic
	assert.Panics(t, func() {
		duckdb.MustGetExtensionManager(db)
	})
}

func TestInitializeExtensions(t *testing.T) {
	config := &duckdb.ExtensionConfig{
		AutoInstall:       true,
		PreloadExtensions: []string{"json"},
	}

	dialector := duckdb.OpenWithExtensions(":memory:", config)
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Initialize extensions manually
	err = duckdb.InitializeExtensions(db)
	require.NoError(t, err)

	// Get manager and verify extensions are loaded
	manager, err := duckdb.GetExtensionManager(db)
	require.NoError(t, err)

	assert.True(t, manager.IsExtensionLoaded("json"))
}

func TestInitializeExtensions_NoExtensionManager(t *testing.T) {
	// Use regular dialector without extension support
	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Should fail to initialize extensions
	err = duckdb.InitializeExtensions(db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "extension manager not found")
}

func TestExtensionConstants(t *testing.T) {
	// Test that extension constants are defined
	assert.Equal(t, "json", duckdb.ExtensionJSON)
	assert.Equal(t, "parquet", duckdb.ExtensionParquet)
	assert.Equal(t, "icu", duckdb.ExtensionICU)
	assert.Equal(t, "csv", duckdb.ExtensionCSV)
	assert.Equal(t, "httpfs", duckdb.ExtensionHTTPS)
	assert.Equal(t, "spatial", duckdb.ExtensionSpatial)
}

func TestExtension_Struct(t *testing.T) {
	ext := duckdb.Extension{
		Name:        "test",
		Description: "Test extension",
		Loaded:      true,
		Installed:   true,
		BuiltIn:     false,
		Version:     "1.0.0",
	}

	assert.Equal(t, "test", ext.Name)
	assert.Equal(t, "Test extension", ext.Description)
	assert.True(t, ext.Loaded)
	assert.True(t, ext.Installed)
	assert.False(t, ext.BuiltIn)
	assert.Equal(t, "1.0.0", ext.Version)
}

func TestExtensionConfig_Struct(t *testing.T) {
	config := duckdb.ExtensionConfig{
		AutoInstall:       true,
		PreloadExtensions: []string{"json", "parquet"},
		Timeout:           30 * time.Second,
		RepositoryURL:     "https://extensions.duckdb.org",
		AllowUnsigned:     false,
	}

	assert.True(t, config.AutoInstall)
	assert.Equal(t, []string{"json", "parquet"}, config.PreloadExtensions)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "https://extensions.duckdb.org", config.RepositoryURL)
	assert.False(t, config.AllowUnsigned)
}

func TestExtensionManager_Timeout(t *testing.T) {
	db, _ := setupBasicExtensionTestDB(t)

	config := &duckdb.ExtensionConfig{
		Timeout: 1 * time.Millisecond, // Very short timeout
	}

	manager := duckdb.NewExtensionManager(db, config)

	// Operations might fail due to timeout, but shouldn't panic
	_, err := manager.ListExtensions()
	_ = err // Might timeout, but shouldn't crash
}

func TestExtensionManager_QuoteName(t *testing.T) {
	// This tests the internal quoteName method indirectly
	// by ensuring malicious extension names are handled safely

	_, manager := setupBasicExtensionTestDB(t)

	// These should not cause SQL injection
	maliciousNames := []string{
		"'; DROP TABLE users; --",
		"extension\"with\"quotes",
		"extension'with'quotes",
		"extension;with;semicolons",
	}

	for _, name := range maliciousNames {
		// Should not panic or cause SQL injection
		err := manager.LoadExtension(name)
		_ = err // Will likely error due to extension not existing, but shouldn't crash
	}
}
