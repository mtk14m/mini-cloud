package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/config"
	"github.com/stretchr/testify/assert"
)

// MockRoundTripper simule une réponse HTTP pour éviter les appels réseau réels.
type MockRoundTripper struct {
	Response *http.Response
	Error    error
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Error
}

func TestLoginValidCredentials(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := &config.Config{
		AuthServiceURL: "http://localhost:8081",
		JWT_SECRET:     "test-secret",
	}

	// Crée une réponse simulée pour le service d'authentification
	authResp := AuthResponse{
		UserID:   "123",
		Username: "testuser",
		Role:     "user",
	}
	jsonResp, _ := json.Marshal(authResp)
	mockResponse := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(jsonResp)),
	}

	// Crée un client HTTP mocké
	mockClient := &http.Client{
		Transport: &MockRoundTripper{Response: mockResponse},
	}

	// Utilise le mock dans le handler
	router.POST("/login", Login(cfg, mockClient))

	// Test data
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}
	jsonData, _ := json.Marshal(loginReq)

	// Test
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
	assert.Contains(t, w.Body.String(), "testuser")
}

func TestLoginInvalidCredentials(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := &config.Config{
		AuthServiceURL: "http://localhost:8081",
		JWT_SECRET:     "test-secret",
	}

	// Crée une réponse simulée en erreur
	mockResponse := &http.Response{
		Status:     "401 Unauthorized",
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"error": "invalid credentials"}`))),
	}

	// Crée un client HTTP mocké
	mockClient := &http.Client{
		Transport: &MockRoundTripper{Response: mockResponse},
	}

	// Utilise le mock dans le handler
	router.POST("/login", Login(cfg, mockClient))

	// Test data
	loginReq := LoginRequest{
		Username: "invaliduser",
		Password: "invalidpass",
	}
	jsonData, _ := json.Marshal(loginReq)

	// Test
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Auth service error")
}

func TestLoginAuthServiceUnreachable(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := &config.Config{
		AuthServiceURL: "http://localhost:8081",
		JWT_SECRET:     "test-secret",
	}

	// Crée un client HTTP qui retourne une erreur de réseau
	mockClient := &http.Client{
		Transport: &MockRoundTripper{Error: fmt.Errorf("connection refused")},
	}

	// Utilise le mock dans le handler
	router.POST("/login", Login(cfg, mockClient))

	// Test data
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}
	jsonData, _ := json.Marshal(loginReq)

	// Test
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Auth service error")
}
