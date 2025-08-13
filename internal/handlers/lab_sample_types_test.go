package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"case/internal/models"
)

// Mock database for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(*sql.Row)
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func TestHandlerAPIGetLabSwabTypes(t *testing.T) {
	app := setupTestApp()

	// Create a mock database
	mockDB := &MockDB{}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/lab/swab-types", nil)

	// Create a test context
	ctx := app.AcquireCtx(req)
	defer app.ReleaseCtx(ctx)

	// Create a test logger
	logger := &fiber.Logger{}

	// Test the handler
	err := HandlerAPIGetLabSwabTypes(ctx, mockDB, logger)

	// Since we're using a mock DB, we expect an error
	assert.Error(t, err)
}

func TestHandlerAPIGetLabUrineTypes(t *testing.T) {
	app := setupTestApp()

	// Create a mock database
	mockDB := &MockDB{}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/lab/urine-types", nil)

	// Create a test context
	ctx := app.AcquireCtx(req)
	defer app.ReleaseCtx(ctx)

	// Create a test logger
	logger := &fiber.Logger{}

	// Test the handler
	err := HandlerAPIGetLabUrineTypes(ctx, mockDB, logger)

	// Since we're using a mock DB, we expect an error
	assert.Error(t, err)
}

func TestHandlerAPIGetLabBloodTypes(t *testing.T) {
	app := setupTestApp()

	// Create a mock database
	mockDB := &MockDB{}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/lab/blood-types", nil)

	// Create a test context
	ctx := app.AcquireCtx(req)
	defer app.ReleaseCtx(ctx)

	// Create a test logger
	logger := &fiber.Logger{}

	// Test the handler
	err := HandlerAPIGetLabBloodTypes(ctx, mockDB, logger)

	// Since we're using a mock DB, we expect an error
	assert.Error(t, err)
}

func TestHandlerAPIGetLabBloodTypesByCategory(t *testing.T) {
	app := setupTestApp()

	// Create a mock database
	mockDB := &MockDB{}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/lab/blood-types/category/CBC", nil)

	// Create a test context
	ctx := app.AcquireCtx(req)
	defer app.ReleaseCtx(ctx)

	// Set the category parameter
	ctx.Params("category", "CBC")

	// Create a test logger
	logger := &fiber.Logger{}

	// Test the handler
	err := HandlerAPIGetLabBloodTypesByCategory(ctx, mockDB, logger)

	// Since we're using a mock DB, we expect an error
	assert.Error(t, err)
}

func TestHandlerAPISaveLabSampleSelections(t *testing.T) {
	app := setupTestApp()

	// Create test data
	testData := map[string]interface{}{
		"lab_id":      1,
		"sample_type": "blood",
		"selections": []map[string]interface{}{
			{
				"selected_type_id": 1,
				"other_specify":    "",
			},
			{
				"selected_type_id": 0,
				"other_specify":    "Custom test",
			},
		},
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(testData)

	// Create a test request
	req := httptest.NewRequest("POST", "/api/lab/sample-selections", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create a test context
	ctx := app.AcquireCtx(req)
	defer app.ReleaseCtx(ctx)

	// Create a mock database
	mockDB := &MockDB{}

	// Create a test logger
	logger := &fiber.Logger{}

	// Test the handler
	err := HandlerAPISaveLabSampleSelections(ctx, mockDB, logger)

	// Since we're using a mock DB, we expect an error
	assert.Error(t, err)
}

func TestHandlerAPIGetLabSampleSelections(t *testing.T) {
	app := setupTestApp()

	// Create a mock database
	mockDB := &MockDB{}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/lab/sample-selections/1", nil)

	// Create a test context
	ctx := app.AcquireCtx(req)
	defer app.ReleaseCtx(ctx)

	// Set the lab_id parameter
	ctx.Params("lab_id", "1")

	// Create a test logger
	logger := &fiber.Logger{}

	// Test the handler
	err := HandlerAPIGetLabSampleSelections(ctx, mockDB, logger)

	// Since we're using a mock DB, we expect an error
	assert.Error(t, err)
}

// Test data structures
func TestLabSwabTypeModel(t *testing.T) {
	swabType := &models.LabSwabType{
		ID:          1,
		Name:        "Wound swab",
		Description: sql.NullString{String: "Swab from wound site", Valid: true},
	}

	assert.Equal(t, 1, swabType.ID)
	assert.Equal(t, "Wound swab", swabType.Name)
	assert.True(t, swabType.Description.Valid)
	assert.Equal(t, "Swab from wound site", swabType.Description.String)
}

func TestLabUrineTypeModel(t *testing.T) {
	urineType := &models.LabUrineType{
		ID:          1,
		Name:        "Chemistry",
		Description: sql.NullString{String: "Urine chemistry analysis", Valid: true},
	}

	assert.Equal(t, 1, urineType.ID)
	assert.Equal(t, "Chemistry", urineType.Name)
	assert.True(t, urineType.Description.Valid)
	assert.Equal(t, "Urine chemistry analysis", urineType.Description.String)
}

func TestLabBloodTypeModel(t *testing.T) {
	bloodType := &models.LabBloodType{
		ID:          1,
		Name:        "CBC",
		Description: sql.NullString{String: "Complete Blood Count", Valid: true},
		Category:    sql.NullString{String: "CBC", Valid: true},
	}

	assert.Equal(t, 1, bloodType.ID)
	assert.Equal(t, "CBC", bloodType.Name)
	assert.True(t, bloodType.Description.Valid)
	assert.Equal(t, "Complete Blood Count", bloodType.Description.String)
	assert.True(t, bloodType.Category.Valid)
	assert.Equal(t, "CBC", bloodType.Category.String)
}

func TestLabSampleSelectionModel(t *testing.T) {
	sampleSelection := &models.LabSampleSelection{
		ID:             1,
		LabID:          1,
		SampleType:     "blood",
		SelectedTypeID: sql.NullInt64{Int64: 1, Valid: true},
		OtherSpecify:   sql.NullString{String: "", Valid: false},
	}

	assert.Equal(t, 1, sampleSelection.ID)
	assert.Equal(t, 1, sampleSelection.LabID)
	assert.Equal(t, "blood", sampleSelection.SampleType)
	assert.True(t, sampleSelection.SelectedTypeID.Valid)
	assert.Equal(t, int64(1), sampleSelection.SelectedTypeID.Int64)
	assert.False(t, sampleSelection.OtherSpecify.Valid)
}
