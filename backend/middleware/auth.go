package middleware

import (
	"net/http"
	"strings"

	"booking-saas/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const ClaimsKey = "auth_claims"

// RequireJWT authenticates the request before any protected handler runs.
func RequireJWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if secret == "" || len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c)
			return
		}

		claims := &utils.JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithIssuer("booking-saas"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
		if err != nil || !token.Valid || claims.UserID <= 0 {
			unauthorized(c)
			return
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// AuthClaims returns claims established by RequireJWT, never client-supplied user data.
func AuthClaims(c *gin.Context) (*utils.JWTClaims, bool) {
	value, exists := c.Get(ClaimsKey)
	claims, ok := value.(*utils.JWTClaims)
	return claims, exists && ok && claims != nil
}

// RequireRole allows any explicitly listed role and must run after RequireJWT.
// No roles means deny all. Business membership must be checked separately.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := AuthClaims(c)
		if !ok {
			unauthorized(c)
			return
		}
		for _, role := range roles {
			if role != "" && claims.Role == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "Forbidden"})
	}
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
}
