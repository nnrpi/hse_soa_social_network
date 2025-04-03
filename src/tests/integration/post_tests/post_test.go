package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	connectionString := "postgres://postgres:postgres@localhost:5432/socialnetwork_test?sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		fmt.Printf("Failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	testDB = db

	setupTestDatabase(testDB)

	exitCode := m.Run()

	teardownTestDatabase(testDB)
	testDB.Close()

	os.Exit(exitCode)
}

type Promo struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Discount    float64 `json:"discount"`
	UsageLimit  int     `json:"usage_limit"`
	UsageCount  int     `json:"usage_count"`
	ValidDays   int     `json:"valid_days"`
}

type CreatePromoRequest struct {
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Discount    float64 `json:"discount"`
	ValidDays   int     `json:"valid_days"`
	UsageLimit  int     `json:"usage_limit"`
}

type ApplyPromoRequest struct {
	UserID string `json:"user_id"`
}

type PromoRepository struct {
	db *sql.DB
}

func NewPromoRepository(db *sql.DB) *PromoRepository {
	return &PromoRepository{db: db}
}

func (r *PromoRepository) GetByCode(ctx context.Context, code string) (*Promo, error) {
	promo := &Promo{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, code, title, description, discount, usage_limit, usage_count FROM promos WHERE code = $1",
		code).Scan(&promo.ID, &promo.Code, &promo.Title, &promo.Description, &promo.Discount,
		&promo.UsageLimit, &promo.UsageCount)
	return promo, err
}

type PromoService struct {
	repo *PromoRepository
}

func NewPromoService(repo *PromoRepository) *PromoService {
	return &PromoService{repo: repo}
}

type PromoHandler struct {
	service *PromoService
}

func NewPromoHandler(service *PromoService) *PromoHandler {
	return &PromoHandler{service: service}
}

func (h *PromoHandler) Create(c *gin.Context) {
	var req CreatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.service.repo.db.Exec(
		"INSERT INTO promos (code, title, description, discount, usage_limit, usage_count) VALUES ($1, $2, $3, $4, $5, 0)",
		req.Code, req.Title, req.Description, req.Discount, req.UsageLimit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Promo code created"})
}

func (h *PromoHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	promo, err := h.service.repo.GetByCode(c, code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Promo code not found"})
		return
	}

	c.JSON(http.StatusOK, promo)
}

func (h *PromoHandler) Apply(c *gin.Context) {
	code := c.Param("code")
	var req ApplyPromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promo, err := h.service.repo.GetByCode(c, code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Promo code not found"})
		return
	}

	if promo.UsageCount >= promo.UsageLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Promo code usage limit reached"})
		return
	}

	_, err = h.service.repo.db.Exec(
		"UPDATE promos SET usage_count = usage_count + 1 WHERE code = $1",
		code)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Promo code applied",
		"discount": promo.Discount,
	})
}

func setupTestDatabase(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS promos")

	db.Exec(`
		CREATE TABLE promos (
			id SERIAL PRIMARY KEY,
			code VARCHAR(50) UNIQUE NOT NULL,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			discount NUMERIC(5,2) NOT NULL,
			usage_limit INTEGER NOT NULL,
			usage_count INTEGER NOT NULL DEFAULT 0
		)
	`)
}

func teardownTestDatabase(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS promos")
}

func setupTestRouter() (*gin.Engine, *PromoRepository) {
	router := gin.Default()

	repo := NewPromoRepository(testDB)
	service := NewPromoService(repo)
	handler := NewPromoHandler(service)

	promoGroup := router.Group("/api/promos")
	{
		promoGroup.POST("/", handler.Create)
		promoGroup.GET("/:code", handler.GetByCode)
		promoGroup.POST("/:code/apply", handler.Apply)
	}

	return router, repo
}

func TestPromoCodeLifecycle(t *testing.T) {
	router, repo := setupTestRouter()

	createPromoData := CreatePromoRequest{
		Code:        "TESTPROMO",
		Title:       "Test Promotion",
		Description: "Test Promo Code",
		Discount:    25.0,
		ValidDays:   30,
		UsageLimit:  5,
	}

	jsonData, _ := json.Marshal(createPromoData)
	req, _ := http.NewRequest("POST", "/api/promos/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin-token")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	createdPromo, err := repo.GetByCode(context.Background(), "TESTPROMO")
	assert.NoError(t, err)
	assert.Equal(t, "TESTPROMO", createdPromo.Code)
	assert.Equal(t, 25.0, createdPromo.Discount)

	applyReq := ApplyPromoRequest{
		UserID: "test-user-1",
	}

	jsonData, _ = json.Marshal(applyReq)
	req, _ = http.NewRequest("POST", "/api/promos/TESTPROMO/apply", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	updatedPromo, _ := repo.GetByCode(context.Background(), "TESTPROMO")
	assert.Equal(t, 1, updatedPromo.UsageCount)

	for i := 0; i < 5; i++ {
		applyReq := ApplyPromoRequest{
			UserID: fmt.Sprintf("test-user-%d", i+2),
		}

		jsonData, _ = json.Marshal(applyReq)
		req, _ = http.NewRequest("POST", "/api/promos/TESTPROMO/apply", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResponse map[string]string
	json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse["error"], "usage limit reached")
}
