package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"unimatch-be/config"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is missing")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
			c.Abort()
			return
		}

		// Set user info to context for downstream handlers
		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Set("is_verified", claims["is_verified"])

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Fail(c, http.StatusForbidden, "FORBIDDEN", "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		isVerifiedRaw, exists := c.Get("is_verified")
		if !exists {
			response.Fail(c, http.StatusForbidden, "FORBIDDEN", "Account verification status unknown")
			c.Abort()
			return
		}
		
		isVerified, ok := isVerifiedRaw.(bool)
		if !ok || !isVerified {
			response.Fail(c, http.StatusForbidden, "FORBIDDEN", "Account not verified. Please wait for admin approval.")
			c.Abort()
			return
		}
		c.Next()
	}
}
