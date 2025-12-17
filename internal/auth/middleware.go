package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const userContextKey contextKey = "godriveUser"

// ContextUser represents the authenticated principal stored in the request context.
type ContextUser struct {
	ID      string
	Email   string
	IsAdmin bool
}

// AuthMiddleware validates bearer tokens and injects the authenticated user.
func AuthMiddleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
			return
		}

		token := extractBearerToken(authHeader)
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := service.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(string(userContextKey), ContextUser{
			ID:      claims.UserID.String(),
			Email:   claims.Email,
			IsAdmin: claims.IsAdmin,
		})

		c.Next()
	}
}

// CurrentUser extracts the authenticated user from the context.
func CurrentUser(c *gin.Context) (ContextUser, bool) {
	value, exists := c.Get(string(userContextKey))
	if !exists {
		return ContextUser{}, false
	}
	user, ok := value.(ContextUser)
	return user, ok
}

// RequireUser fetches the authenticated user and parses the identifier.
func RequireUser(c *gin.Context) (uuid.UUID, ContextUser, bool) {
	user, ok := CurrentUser(c)
	if !ok {
		return uuid.Nil, ContextUser{}, false
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return uuid.Nil, ContextUser{}, false
	}
	return id, user, true
}

func extractBearerToken(header string) string {
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return ""
	}
	return strings.TrimSpace(header[7:])
}
