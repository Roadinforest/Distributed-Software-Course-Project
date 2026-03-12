package service

import (
	"errors"
	"fmt"
	"time"

	"user-service/internal/model"
	jwtpkg "user-service/internal/pkg/jwt"
	"user-service/internal/pkg/security"
	"user-service/internal/repository"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrNothingToUpdate    = errors.New("nothing to update")
)

type AuthService interface {
	Register(req model.RegisterRequest) (*model.User, error)
	Login(req model.LoginRequest) (string, time.Time, *model.User, error)
	GetProfile(userID uint64) (*model.User, error)
	UpdateProfile(userID uint64, req model.UpdateProfileRequest) error
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *jwtpkg.Manager
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *jwtpkg.Manager) AuthService {
	return &authService{userRepo: userRepo, jwtManager: jwtManager}
}

func (s *authService) Register(req model.RegisterRequest) (*model.User, error) {
	existing, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	user := &model.User{
		Username: req.Username,
		Password: hashed,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(req model.LoginRequest) (string, time.Time, *model.User, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	if user == nil || !security.VerifyPassword(user.Password, req.Password) {
		return "", time.Time{}, nil, ErrInvalidCredentials
	}
	if user.Status == 0 {
		return "", time.Time{}, nil, ErrInvalidCredentials
	}

	token, expiresAt, err := s.jwtManager.Generate(user.ID, user.Username)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	return token, expiresAt, user, nil
}

func (s *authService) GetProfile(userID uint64) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *authService) UpdateProfile(userID uint64, req model.UpdateProfileRequest) error {
	if req.Email == "" && req.Phone == "" {
		return ErrNothingToUpdate
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	email := user.Email
	phone := user.Phone
	if req.Email != "" {
		email = req.Email
	}
	if req.Phone != "" {
		phone = req.Phone
	}

	return s.userRepo.UpdateProfile(userID, email, phone)
}
