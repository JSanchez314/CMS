package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"JSanches/CMD/database"
	"JSanches/CMD/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// --- Interfaces para inyección de dependencias ---

type PostgresRepository interface {
	CreateContent(content *models.Content) error
	FindContents(contents *[]models.Content) error
	GetContent(id string, content *models.Content) error
	SaveContent(content *models.Content) error
	DeleteContent(content *models.Content) error
}

type MongoRepository interface {
	InsertContentBody(ctx context.Context, contentBody *models.ContentBody) error
	GetContentBody(ctx context.Context, filter interface{}, contentBody *models.ContentBody) error
	UpdateContentBody(ctx context.Context, filter interface{}, update interface{}) error
	DeleteContentBody(ctx context.Context, filter interface{}) error
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

// --- Handler con dependencias inyectadas ---

type ContentHandler struct {
	PG    PostgresRepository
	Mongo MongoRepository
	Cache CacheRepository
}

func NewContentHandler(pg PostgresRepository, mongo MongoRepository, cache CacheRepository) *ContentHandler {
	return &ContentHandler{
		PG:    pg,
		Mongo: mongo,
		Cache: cache,
	}
}

// CreateContent crea un nuevo contenido
func (h *ContentHandler) CreateContent(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not authorized"})
	}

	type CreateContentRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	var req CreateContentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	content := models.Content{
		Title:       req.Title,
		Description: req.Description,
		UserID:      int(userID),
	}

	if err := h.PG.CreateContent(&content); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create content"})
	}

	contentBody := models.ContentBody{
		ContentID: content.ID,
		Body:      req.Description,
	}
	if err := h.Mongo.InsertContentBody(context.Background(), &contentBody); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create content body in MongoDB"})
	}

	return c.Status(fiber.StatusCreated).JSON(content)
}

// ListContents obtiene la lista de contenidos
func (h *ContentHandler) ListContents(c *fiber.Ctx) error {
	cacheKey := "contents_list"
	if cached, err := h.Cache.Get(context.Background(), cacheKey); err == nil && cached != "" {
		return c.Status(fiber.StatusOK).JSON(cached)
	}

	var contents []models.Content
	if err := h.PG.FindContents(&contents); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve contents"})
	}

	_ = h.Cache.Set(context.Background(), cacheKey, fmt.Sprintf("%v", contents), 10*time.Second)
	return c.Status(fiber.StatusOK).JSON(contents)
}

// GetContent obtiene un contenido por su id
func (h *ContentHandler) GetContent(c *fiber.Ctx) error {
	id := c.Params("id")
	var content models.Content
	if err := h.PG.GetContent(id, &content); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Content not found"})
	}

	// Convertir id a entero para el filtro de Mongo
	contentID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid content ID"})
	}
	filter := bson.M{"content_id": contentID}
	var contentBody models.ContentBody
	if err := h.Mongo.GetContentBody(context.Background(), filter, &contentBody); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Content body not found"})
	}

	response := fiber.Map{
		"id":          content.ID,
		"title":       content.Title,
		"author":      content.UserID,
		"description": contentBody.Body,
		"created_at":  content.CreatedAt,
		"updated_at":  content.UpdatedAt,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateContent actualiza un contenido por su id
func (h *ContentHandler) UpdateContent(c *fiber.Ctx) error {
	id := c.Params("id")
	var content models.Content
	if err := h.PG.GetContent(id, &content); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Content not found"})
	}

	type UpdateContentRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	var req UpdateContentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	content.Title = req.Title
	content.Description = req.Description

	if err := h.PG.SaveContent(&content); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save content"})
	}

	contentID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid content ID"})
	}
	filter := bson.M{"content_id": contentID}
	update := bson.M{
		"$set": bson.M{
			"body":  req.Description,
			"title": req.Title,
		},
	}
	if err := h.Mongo.UpdateContentBody(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update MongoDB"})
	}

	return c.Status(fiber.StatusOK).JSON(content)
}

// DeleteContent elimina un contenido por su id
func (h *ContentHandler) DeleteContent(c *fiber.Ctx) error {
	id := c.Params("id")
	var content models.Content
	if err := h.PG.GetContent(id, &content); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Content not found"})
	}
	if err := h.PG.DeleteContent(&content); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete content"})
	}

	contentID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid content ID"})
	}
	filter := bson.M{"content_id": contentID}
	_ = h.Mongo.DeleteContentBody(context.Background(), filter)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Content deleted successfully"})
}

type PGClient struct{}

func (p *PGClient) CreateContent(content *models.Content) error {
	result := database.PostgresGetDB().Create(content)
	return result.Error
}

func (p *PGClient) FindContents(contents *[]models.Content) error {
	result := database.PostgresGetDB().Find(contents)
	return result.Error
}

func (p *PGClient) GetContent(id string, content *models.Content) error {
	result := database.PostgresGetDB().First(content, id)
	return result.Error
}

func (p *PGClient) SaveContent(content *models.Content) error {
	result := database.PostgresGetDB().Save(content)
	return result.Error
}

func (p *PGClient) DeleteContent(content *models.Content) error {
	result := database.PostgresGetDB().Delete(content)
	return result.Error
}

// MongoClient implementa MongoRepository usando la base de datos Mongo real
type MongoClient struct{}

func (m *MongoClient) InsertContentBody(ctx context.Context, contentBody *models.ContentBody) error {
	_, err := database.MongoGetCollection("contents", "contents").InsertOne(ctx, contentBody)
	return err
}

func (m *MongoClient) GetContentBody(ctx context.Context, filter interface{}, contentBody *models.ContentBody) error {
	return database.MongoGetCollection("contents", "contents").FindOne(ctx, filter).Decode(contentBody)
}

func (m *MongoClient) UpdateContentBody(ctx context.Context, filter interface{}, update interface{}) error {
	_, err := database.MongoGetCollection("contents", "contents").UpdateMany(ctx, filter, update)
	return err
}

func (m *MongoClient) DeleteContentBody(ctx context.Context, filter interface{}) error {
	_, err := database.MongoGetCollection("contents", "contents").DeleteOne(ctx, filter)
	return err
}

// RedisClient implementa CacheRepository usando el cliente Redis real
type RedisClient struct{}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return database.RedisGetClient().Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return database.RedisGetClient().Set(ctx, key, value, expiration).Err()
}
