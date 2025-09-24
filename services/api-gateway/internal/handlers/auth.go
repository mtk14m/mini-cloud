package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/config"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login gère la connexion de l'utilisateur.
func Login(cfg *config.Config, client ...*http.Client) gin.HandlerFunc {
	httpClient := http.DefaultClient
	if len(client) > 0 {
		httpClient = client[0]
	}

	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Appeler le service d'authentification
		authResp, err := callAuthService(httpClient, cfg.AuthServiceURL+"/login", req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service error: " + err.Error()})
			return
		}

		// Générer un token JWT
		token, err := generateJWT(authResp.UserID, authResp.Username, authResp.Role, cfg.JWT_SECRET)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":    token,
			"user_id":  authResp.UserID,
			"username": authResp.Username,
			"role":     authResp.Role,
		})
	}
}

// Register gère l'inscription d'un nouvel utilisateur.
func Register(cfg *config.Config, client ...*http.Client) gin.HandlerFunc {
	httpClient := http.DefaultClient
	if len(client) > 0 {
		httpClient = client[0]
	}

	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return // Ajout du return manquant
		}

		// Appeler le service d'authentification
		authResp, err := callAuthService(httpClient, cfg.AuthServiceURL+"/register", req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service error: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user_id": authResp.UserID,
		})
	}
}

// Validate valide le token JWT et retourne les informations de l'utilisateur.
func Validate(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Le middleware Auth a déjà validé le token
		userID := c.GetString("user_id")
		username := c.GetString("username")
		role := c.GetString("role")

		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	}
}

// callAuthService appelle le service d'authentification externe.
func callAuthService(client *http.Client, url string, data interface{}) (*AuthResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned non-200 status: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(body))
	}

	return &authResp, nil
}

// generateJWT génère un token JWT pour l'utilisateur.
func generateJWT(userID, username, role, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
