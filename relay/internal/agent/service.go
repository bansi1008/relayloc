package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) NewAgent(
	ctx context.Context,
	Name string,
	id uuid.UUID,
	Token string,
) (*Agent, error) {

	if Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	if id == uuid.Nil {
		return nil, fmt.Errorf("user id is required")
	}

	agent := &Agent{
		ID:              uuid.New(),
		UserID:          id,
		Name:            Name,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		LastConnectedAt: nil,
		Connected:       false,
		Token:           Token,
	}

	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	return agent, nil
}

func (s *Service) GetAgentByID(
	ctx context.Context,
	id uuid.UUID,
) (*Agent, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("agent id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAgentsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*Agent, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) DeleteAgent(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return fmt.Errorf("agent id is required")
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetAgentByName(
	ctx context.Context,
	name string,
	userID uuid.UUID,
) (*Agent, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.GetByName(ctx, name, userID)
}

func (s *Service) AuthenticateAgent(
	ctx context.Context,
	name string,
	rawToken string,
) (*Agent, error) {
	if name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if rawToken == "" {
		return nil, fmt.Errorf("token is required")
	}

	agents, err := s.repo.GetByNameOnly(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	if len(agents) == 0 {
		return nil, fmt.Errorf("agent not found")
	}

	for _, a := range agents {
		if err := bcrypt.CompareHashAndPassword([]byte(a.Token), []byte(rawToken)); err == nil {
			_ = s.repo.UpdateLastConnected(ctx, a.ID)
			return a, nil
		}
	}

	return nil, fmt.Errorf("invalid token")
}
