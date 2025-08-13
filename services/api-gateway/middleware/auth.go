package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/lucas/shared/utils"
    "github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	jwtSecret   []byte
	redisClient *redis.Client
}

func NewAuthMiddleware() *AuthMiddleware {
	secret := utils.GetEnvOrDefault("JWT_SECRECT", "your-secret-key")
	redisAddr := utils.GetEnvOrDefault("REDIS_ADDR", "localhost:6379")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &AuthMiddleware{
		jwtSecret: []byte(secret),
		redisClient: rdb,
	}
}

func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Authorization header required"})
			c.Abort()
			return
		}

		// Validate Bearer token format
		tokenString := strings.TrimPrefix(authHeader, "Bearer")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid authorization format"})
			c.Abort()
			return
		}

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return a.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid token claims"})
			c.Abort()
			return
		}

		// Check if the token is blacklisted (logout/revoked tokens)
		userID := claims["user_id"].(string)
		jti := claims["jti"].(string) // JWT ID for blacklisting specific tokens

		isBlacklisted, err := a.redisClient.Get(context.Background(), "blacklist:"+jti).Result()
		if err != nil && isBlacklisted == "true" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Set user contact for downstream handlers
		c.Set("user_id", userID)
		c.Set("user_email", claims["email"])
		c.Set("user_roles", claims["roles"])

		c.Next()
	}
}

func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Next()
            return
        }

        // Same validation logic but don't abort on failure
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return a.jwtSecret, nil
        })

        if err == nil && token.Valid {
            if claims, ok := token.Claims.(jwt.MapClaims); ok {
                c.Set("user_id", claims["user_id"])
                c.Set("user_email", claims["email"])
                c.Set("user_roles", claims["roles"])
            }
        }

        c.Next()
    }
}
