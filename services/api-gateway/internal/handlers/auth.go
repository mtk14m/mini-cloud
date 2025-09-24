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

func Login(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//Appeler le service d'authentification
		authResp, err := callAuthservice(cfg.AuthServiceURL+"/login", req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service error"})
			return
		}

		//Générer un token JWT
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

func Register(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		//Appeler le service d'authentification
		authResp, err := callAuthservice(cfg.AuthServiceURL+"/register", req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created succesfully",
			"user_id": authResp.UserID,
		})
	}
}

func Validate(cfg *config.Config) gin.HandlerFunc {

	return func(c *gin.Context) {
		//Le middleware Auth à déja validé le token
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

// Fonction utilitaire pour appeler le service d'authentification
func callAuthservice(url string, data interface{}) (*AuthResponse, error) {

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to authenticate: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

// Fonction utilitaire pour générer un token JWT
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
