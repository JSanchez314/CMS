package services

import (
	"JSanches/CMD/models"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Interfaces para inyección de dependencias

// Estructura que implementa UserRepository
type PostgresUserRepository struct {
	DB *gorm.DB
}

// Constructor que devuelve una instancia de UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{DB: db}
}

func (r *PostgresUserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *PostgresUserRepository) FindByUsernameOrEmail(username, email string, user *models.User) error {
	return r.DB.Where("username = ? OR email = ?", username, email).First(user).Error
}

type UserRepository interface {
	Create(user *models.User) error
	// En Login se usa para buscar el usuario:
	FindByUsernameOrEmail(username, email string, user *models.User) error
}

type TokenService interface {
	CreateNewAuthToken(userID string, username string) (string, error)
}

type JWTService struct {
	SecretKey string
}

func NewTokenService(secret string) TokenService {
	return &JWTService{SecretKey: secret}
}

func (s *JWTService) CreateNewAuthToken(userID string, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.SecretKey))
}

// Solicitudes (requests) que ya tenés definidas

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Handler que contiene las dependencias (repositorio y servicio de tokens)

type Handler struct {
	UserRepo     UserRepository
	TokenService TokenService
}

func (h *Handler) RegisterHandler(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	user := models.User{
		Username:          req.Username,
		Email:             req.Email,
		Password:          string(hashedPassword),
		IsActive:          true,
		IsBanned:          false,
		FailedAttempts:    0,
		LastFailedAttempt: time.Time{},
		LockoutUntil:      nil,
	}

	if err := h.UserRepo.Create(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Asumimos que la base de datos asigna un ID válido al usuario (por ejemplo, 123)
	signedToken, err := h.TokenService.CreateNewAuthToken(fmt.Sprintf("%d", user.ID), user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create token",
		})
	}

	cookie := fiber.Cookie{
		Name:     "auth_token",
		Value:    signedToken,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		Secure:   true,
	}
	c.Cookie(&cookie)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User created successfully",
	})
}

func (h *Handler) LoginHandler(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user := models.User{}
	if err := h.UserRepo.FindByUsernameOrEmail(req.Username, req.Email, &user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if user.IsBanned {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User is banned",
		})
	}

	if user.FailedAttempts >= 5 {
		if user.LockoutUntil != nil && user.LockoutUntil.After(time.Now()) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Account is locked",
			})
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	signedToken, err := h.TokenService.CreateNewAuthToken(fmt.Sprintf("%d", user.ID), user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create token",
		})
	}
	cookie := fiber.Cookie{
		Name:     "auth_token",
		Value:    signedToken,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		Secure:   true,
	}
	c.Cookie(&cookie)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
	})
}
