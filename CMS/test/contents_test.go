package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"JSanches/CMD/models"
	"JSanches/CMD/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks para cada interfaz ---

type MockPGRepository struct {
	mock.Mock
}

func (m *MockPGRepository) CreateContent(content *models.Content) error {
	args := m.Called(content)
	return args.Error(0)
}

func (m *MockPGRepository) FindContents(contents *[]models.Content) error {
	args := m.Called(contents)
	return args.Error(0)
}

func (m *MockPGRepository) GetContent(id string, content *models.Content) error {
	args := m.Called(id, content)
	return args.Error(0)
}

func (m *MockPGRepository) SaveContent(content *models.Content) error {
	args := m.Called(content)
	return args.Error(0)
}

func (m *MockPGRepository) DeleteContent(content *models.Content) error {
	args := m.Called(content)
	return args.Error(0)
}

type MockMongoRepository struct {
	mock.Mock
}

func (m *MockMongoRepository) InsertContentBody(ctx context.Context, contentBody *models.ContentBody) error {
	args := m.Called(ctx, contentBody)
	return args.Error(0)
}

func (m *MockMongoRepository) GetContentBody(ctx context.Context, filter interface{}, contentBody *models.ContentBody) error {
	args := m.Called(ctx, filter, contentBody)
	return args.Error(0)
}

func (m *MockMongoRepository) UpdateContentBody(ctx context.Context, filter interface{}, update interface{}) error {
	args := m.Called(ctx, filter, update)
	return args.Error(0)
}

func (m *MockMongoRepository) DeleteContentBody(ctx context.Context, filter interface{}) error {
	args := m.Called(ctx, filter)
	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func TestCreateContent(t *testing.T) {
	// Crear una nueva aplicación Fiber para test
	app := fiber.New()

	// Crear mocks y el handler de contenido
	pgMock := new(MockPGRepository)
	mongoMock := new(MockMongoRepository)
	cacheMock := new(MockCacheRepository)
	handler := services.NewContentHandler(pgMock, mongoMock, cacheMock)

	// Registrar la ruta de prueba para CreateContent
	app.Post("/content", handler.CreateContent)

	// Agregar un middleware que simule un "userId" en los Locals (necesario para el handler)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", float64(1))
		return c.Next()
	})

	// Preparar request
	reqBody := map[string]string{
		"title":       "Test Title",
		"description": "Test Description",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/content", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Configurar expectativas en las mocks
	pgMock.On("CreateContent", mock.AnythingOfType("*models.Content")).Return(nil)
	mongoMock.On("InsertContentBody", mock.Anything, mock.AnythingOfType("*models.ContentBody")).Return(nil)

	// Ejecutar la petición usando Fiber (esto crea internamente el contexto)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	pgMock.AssertExpectations(t)
	mongoMock.AssertExpectations(t)
}

func TestGetContent(t *testing.T) {
	app := fiber.New()

	// Crear mocks y el handler de contenido
	pgMock := new(MockPGRepository)
	mongoMock := new(MockMongoRepository)
	cacheMock := new(MockCacheRepository)
	handler := services.NewContentHandler(pgMock, mongoMock, cacheMock)

	// Registrar la ruta de prueba para GetContent, aprovechando parámetro en la URL
	app.Get("/content/:id", handler.GetContent)

	// Configurar el contenido esperado en PG
	contentID := "1"
	expectedContent := models.Content{
		Title:       "Test Title",
		Description: "Test Description",
		UserID:      1,
	}
	pgMock.On("GetContent", contentID, mock.AnythingOfType("*models.Content")).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*models.Content)
		*arg = expectedContent
	}).Return(nil)

	// Configurar el cuerpo del contenido esperado en Mongo
	filter := map[string]interface{}{"content_id": 1}
	expectedContentBody := models.ContentBody{
		ContentID: 1,
		Body:      "Test Description",
	}
	mongoMock.On("GetContentBody", mock.Anything, filter, mock.AnythingOfType("*models.ContentBody")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*models.ContentBody)
		*arg = expectedContentBody
	})

	// Realizar la petición
	req := httptest.NewRequest("GET", "/content/1", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	pgMock.AssertExpectations(t)
	mongoMock.AssertExpectations(t)
}
