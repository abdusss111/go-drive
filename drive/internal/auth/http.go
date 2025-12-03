package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts authentication endpoints under /auth.
func RegisterRoutes(router *gin.RouterGroup, service *Service) {
	handler := &httpHandler{service: service}
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handler.register)
		authGroup.POST("/login", handler.login)
	}
}

type httpHandler struct {
	service *Service
}

type registerRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Password    string  `json:"password" binding:"required,min=8,max=72"`
	DisplayName *string `json:"display_name" binding:"omitempty,max=128"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type authResponse struct {
	User struct {
		ID          string     `json:"id"`
		Email       string     `json:"email"`
		DisplayName *string    `json:"display_name,omitempty"`
		IsAdmin     bool       `json:"is_admin"`
		CreatedAt   *time.Time `json:"created_at,omitempty"`
	} `json:"user"`
	Tokens struct {
		AccessToken        string `json:"access_token"`
		AccessTokenExpiry  int64  `json:"access_token_expires_at"`
		RefreshToken       string `json:"refresh_token"`
		RefreshTokenExpiry int64  `json:"refresh_token_expires_at"`
	} `json:"tokens"`
}

func (h *httpHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Register(c.Request.Context(), RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		switch err {
		case ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		case ErrInvalidCredentials:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		default:
			// Include error message for debugging
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to register user",
				"detail": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, marshalAuthResponse(result))
}

func (h *httpHandler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			// Include error message for debugging
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to authenticate",
				"detail": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, marshalAuthResponse(result))
}

func marshalAuthResponse(result AuthResult) authResponse {
	resp := authResponse{}
	resp.User.ID = result.User.ID.String()
	resp.User.Email = result.User.Email
	resp.User.DisplayName = result.User.DisplayName
	resp.User.IsAdmin = result.User.IsAdmin
	if !result.User.CreatedAt.IsZero() {
		created := result.User.CreatedAt.UTC()
		resp.User.CreatedAt = &created
	}
	resp.Tokens.AccessToken = result.Tokens.AccessToken
	resp.Tokens.RefreshToken = result.Tokens.RefreshToken
	resp.Tokens.AccessTokenExpiry = result.Tokens.AccessTokenExpiry.Unix()
	resp.Tokens.RefreshTokenExpiry = result.Tokens.RefreshTokenExpiry.Unix()
	return resp
}
