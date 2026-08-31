package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo         Repository
	tokenService *TokenService
}

func NewService(repo Repository, tokenService *TokenService) *Service {
	return &Service{
		repo:         repo,
		tokenService: tokenService,
	}
}

func (s *Service) Register(
	ctx context.Context,
	name string,
	email string,
	password string,
) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if password == "" {
		return nil, fmt.Errorf("password hash is required")
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user already exists")
	}
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	return user, nil
}

func (s *Service) Loginserivce(
	ctx context.Context,
	Email string,
	Password string,
) (*loginResponse, error) {
	Email = strings.ToLower(strings.TrimSpace(Email))
	if Email == " " {
		return nil, fmt.Errorf("email is require")
	}
	if Password == " " {
		return nil, fmt.Errorf("passowrd is require")
	}

	ex, err := s.repo.GetByEmail(ctx, Email)

	if err == nil && ex == nil {
		return nil, fmt.Errorf("please create an account")
	}

	//
	err = bcrypt.CompareHashAndPassword(
		[]byte(ex.PasswordHash),
		[]byte(Password),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	jwt, err := s.tokenService.Generate(ex.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	loginResponse := &loginResponse{
		Token:     jwt,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	return loginResponse, nil
}

type loginResponse struct {
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	return s.repo.GetByEmail(ctx, email)
}
