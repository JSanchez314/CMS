package test

import (
	"JSanches/CMD/models"
	"JSanches/CMD/services"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsernameOrEmail(username, email string, user *models.User) error {
	args := m.Called(username, email)
	// Si se simula un resultado, asignamos los datos al parámetro user.
	if u, ok := args.Get(0).(*models.User); ok {
		*user = *u
	}
	return args.Error(1)
}

// Mock para TokenService
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) CreateNewAuthToken(userID, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func TestRegisterHandler_WithMocks(t *testing.T) {
	// Crear mocks
	userRepoMock := new(MockUserRepository)
	tokenSvcMock := new(MockTokenService)

	// Instanciar el handler con las dependencias inyectadas (mocks)
	h := &services.Handler{
		UserRepo:     userRepoMock,
		TokenService: tokenSvcMock,
	}

	app := fiber.New()
	app.Post("/register", h.RegisterHandler)

	// Datos de la solicitud
	reqData := services.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "securepassword",
	}
	body, _ := json.Marshal(reqData)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Configurar expectativas:
	// Cuando se invoque Create, simulamos éxito y asignamos un ID (p.ej.: 123) al usuario
	userRepoMock.
		On("Create", mock.AnythingOfType("*models.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(0).(*models.User)
			user.ID = 123
		})
	// Cuando se llame al servicio de tokens, retornar un token simulado
	tokenSvcMock.
		On("CreateNewAuthToken", "123", "testuser").
		Return("faketoken", nil)

	// Ejecutar la solicitud
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verificar el contenido de la respuesta
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "User created successfully", respBody["message"])

	// Verificar que los mocks se llamaron según lo esperado
	userRepoMock.AssertExpectations(t)
	tokenSvcMock.AssertExpectations(t)
}

func TestLoginHandler_WithMocks(t *testing.T) {
	// Crear mocks
	userRepoMock := new(MockUserRepository)
	tokenSvcMock := new(MockTokenService)

	// Instanciar el handler con las dependencias inyectadas (mocks)
	h := &services.Handler{
		UserRepo:     userRepoMock,
		TokenService: tokenSvcMock,
	}

	app := fiber.New()
	app.Post("/login", h.LoginHandler)

	// Preparar un usuario ficticio (simula que la base de datos encuentra al usuario)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("rightpassword"), bcrypt.DefaultCost)
	fakeUser := &models.User{
		ID:             456,
		Username:       "testuser",
		Email:          "test@example.com",
		Password:       string(hashedPassword),
		IsBanned:       false,
		FailedAttempts: 0,
		LockoutUntil:   nil,
	}

	// Datos de la solicitud de login
	reqData := services.LoginRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "rightpassword",
	}
	body, _ := json.Marshal(reqData)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Configurar expectativas:
	// Cuando se invoque FindByUsernameOrEmail, retornar el usuario simulado
	userRepoMock.
		On("FindByUsernameOrEmail", "testuser", "test@example.com").
		Return(fakeUser, nil)
	// Cuando se llame al servicio de tokens, retornar un token simulado
	tokenSvcMock.
		On("CreateNewAuthToken", fmt.Sprintf("%d", fakeUser.ID), fakeUser.Username).
		Return("faketoken", nil)

	// Ejecutar la solicitud
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verificar el contenido de la respuesta
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Login successful", respBody["message"])

	// Verificar que los mocks se llamaron según lo esperado
	userRepoMock.AssertExpectations(t)
	tokenSvcMock.AssertExpectations(t)
}
