package service

import (
	"context"
	"testing"
	"time"

	"coursework/auth-service/internal/domain"
)

type stubUserRepo struct{}

func (stubUserRepo) CreateUser(ctx context.Context, email, fullName, role, passwordHash string) (domain.User, error) {
	return domain.User{
		Email:    email,
		FullName: fullName,
		Role:     role,
	}, nil
}

func (stubUserRepo) GetByEmail(context.Context, string) (domain.User, error) {
	return domain.User{}, nil
}

func (stubUserRepo) Ping(context.Context) error {
	return nil
}

func TestRegisterAllowsClientOnly(t *testing.T) {
	svc := NewAuthService(stubUserRepo{}, "secret", time.Hour)

	if _, err := svc.Register(context.Background(), "client@example.com", "Pass1234", "Client User", "client"); err != nil {
		t.Fatalf("expected client registration to succeed, got %v", err)
	}

	if _, err := svc.Register(context.Background(), "manager@example.com", "Pass1234", "Manager User", "manager"); err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole for manager registration, got %v", err)
	}
}

func TestRegisterValidatesInput(t *testing.T) {
	svc := NewAuthService(stubUserRepo{}, "secret", time.Hour)

	if _, err := svc.Register(context.Background(), "invalid-email", "Pass1234", "Client User", "client"); err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
	if _, err := svc.Register(context.Background(), "client@example.com", "short", "Client User", "client"); err != ErrWeakPassword {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
	if _, err := svc.Register(context.Background(), "client@example.com", "Pass1234", "   ", "client"); err != ErrInvalidFullName {
		t.Fatalf("expected ErrInvalidFullName, got %v", err)
	}
}
