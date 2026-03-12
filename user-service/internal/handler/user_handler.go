package handler

import (
	"errors"
	"net/http"
	"time"

	"user-service/internal/middleware"
	"user-service/internal/model"
	"user-service/internal/pkg/response"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService service.AuthService
}

func NewUserHandler(authService service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, "invalid request params", nil)
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.JSON(c, http.StatusBadRequest, "username already exists", nil)
			return
		}
		response.JSON(c, http.StatusInternalServerError, "register failed", nil)
		return
	}

	response.JSON(c, http.StatusOK, "register success", gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"phone":      user.Phone,
		"created_at": user.CreatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, "invalid request params", nil)
		return
	}

	token, expiresAt, user, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.JSON(c, http.StatusUnauthorized, "invalid username or password", nil)
			return
		}
		response.JSON(c, http.StatusInternalServerError, "login failed", nil)
		return
	}

	response.JSON(c, http.StatusOK, "login success", gin.H{
		"token":      token,
		"expires_at": expiresAt.UTC().Format(time.RFC3339),
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.JSON(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	user, err := h.authService.GetProfile(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.JSON(c, http.StatusNotFound, "user not found", nil)
			return
		}
		response.JSON(c, http.StatusInternalServerError, "query profile failed", nil)
		return
	}

	response.JSON(c, http.StatusOK, "success", gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"phone":      user.Phone,
		"created_at": user.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": user.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.JSON(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, "invalid request params", nil)
		return
	}

	if err := h.authService.UpdateProfile(userID, req); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.JSON(c, http.StatusNotFound, "user not found", nil)
			return
		}
		if errors.Is(err, service.ErrNothingToUpdate) {
			response.JSON(c, http.StatusBadRequest, "nothing to update", nil)
			return
		}
		response.JSON(c, http.StatusInternalServerError, "update profile failed", nil)
		return
	}

	response.JSON(c, http.StatusOK, "update success", nil)
}
