package agent

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"relaygo/relay/internal/user"
)

type Service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, userRepo user.Repository) *Service {
	return &Service{
		repo:     repo,
		userRepo: userRepo,
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

func (s *Service) AuthenticateWithKey(
	ctx context.Context,
	email string,
	name string,
	rawToken string,
	publicKey string,
) (*Agent, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	rawToken = strings.TrimSpace(rawToken)

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if rawToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	a, err := s.repo.GetByName(ctx, name, u.ID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	if a.PublicKey != "" {
		return nil, fmt.Errorf("agent is already activated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(a.Token), []byte(rawToken)); err != nil {
		return nil, fmt.Errorf("invalid access token")
	}

	if publicKey != "" {
		if err := s.repo.UpdatePublicKey(ctx, a.ID, publicKey); err != nil {
			return nil, fmt.Errorf("update public key: %w", err)
		}
	}

	_ = s.repo.UpdateLastConnected(ctx, a.ID)
	return a, nil
}

func (s *Service) CreateChallenge(ctx context.Context, agentIDStr string) (*Agent, string, error) {
	agentID, err := uuid.Parse(strings.TrimSpace(agentIDStr))
	if err != nil {
		return nil, "", fmt.Errorf("invalid agent id")
	}

	a, err := s.repo.GetByID(ctx, agentID)
	if err != nil {
		return nil, "", fmt.Errorf("agent not found")
	}

	if a.PublicKey == "" {
		return nil, "", fmt.Errorf("please activate the agent first")
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, "", fmt.Errorf("generate nonce: %w", err)
	}

	nonce := hex.EncodeToString(b)
	return a, nonce, nil
}

func (s *Service) VerifyChallenge(ctx context.Context, agentIDStr string, nonce string, signatureBase64 string) (*Agent, error) {
	agentID, err := uuid.Parse(strings.TrimSpace(agentIDStr))
	if err != nil {
		return nil, fmt.Errorf("invalid agent id")
	}

	if nonce == "" {
		return nil, fmt.Errorf("invalid challenge session")
	}

	a, err := s.repo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	if a.PublicKey == "" {
		return nil, fmt.Errorf("please activate the agent first")
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(a.PublicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid stored public key")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid signature format")
	}

	if !ed25519.Verify(pubKeyBytes, []byte(nonce), sigBytes) {
		return nil, fmt.Errorf("signature verification failed")
	}

	_ = s.repo.UpdateLastConnected(ctx, a.ID)
	return a, nil
}


