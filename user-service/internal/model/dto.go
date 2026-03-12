package model

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=128"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email,max=128"`
	Phone string `json:"phone" binding:"omitempty,max=20"`
}
