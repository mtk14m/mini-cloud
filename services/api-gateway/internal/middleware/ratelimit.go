package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(redisURL string, limit int) *RateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	return &RateLimiter{
		redis:  rdb,
		limit:  limit,
		window: time.Minute, // Fenêtre de 1 minute
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer l'IP du client
		clientIP := c.ClientIP()

		// Clé Redis pour cette IP
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Incrémenter le compteur
		count, err := rl.redis.Incr(c.Request.Context(), key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			c.Abort()
			return
		}

		// Si c'est la première requête, définir l'expiration
		if count == 1 {
			rl.redis.Expire(c.Request.Context(), key, rl.window)
		}

		// Vérifier la limite
		if count > int64(rl.limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":  "Rate limit exceeded",
				"limit":  rl.limit,
				"window": rl.window.String(),
			})
			c.Abort()
			return
		}

		// Ajouter les headers de rate limiting
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.limit-int(count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.window).Unix()))

		c.Next()
	}
}
