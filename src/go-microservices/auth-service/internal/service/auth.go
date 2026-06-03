package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"coursework/auth-service/internal/domain"
	"coursework/auth-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password must be at least 8 characters long")
	ErrInvalidFullName    = errors.New("full name is required")
)

type userRepo interface {
	CreateUser(ctx context.Context, email, fullName, role, passwordHash string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Ping(ctx context.Context) error
}

type AuthService struct {
	repo      userRepo //что будет, если userRepo не будет существовать. AuthService не может существовать. Тогда какая это связь будет? Изучить.
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(repo userRepo, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

func (s *AuthService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName, role string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	fullName = strings.TrimSpace(fullName)
	role = strings.ToLower(strings.TrimSpace(role))

	if !strings.Contains(email, "@") {
		return domain.User{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return domain.User{}, ErrWeakPassword
	}
	if fullName == "" {
		return domain.User{}, ErrInvalidFullName
	}
	// Public self-registration is intentionally limited to client accounts.
	if role != "client" {
		return domain.User{}, ErrInvalidRole
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	return s.repo.CreateUser(ctx, email, fullName, role, string(hash))
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.User{}, ErrInvalidCredentials
		}
		return "", domain.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.User{}, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
		"exp":       time.Now().Add(s.tokenTTL).Unix(),
		"iat":       time.Now().Unix(),
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, user, nil
}

func IsEmailExists(err error) bool {
	return errors.Is(err, repository.ErrEmailExists)
}
