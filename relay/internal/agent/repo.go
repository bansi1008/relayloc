package agent

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, agent *Agent) error
	GetByID(ctx context.Context, id uuid.UUID) (*Agent, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Agent, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string, userID uuid.UUID) (*Agent, error)
	GetByNameOnly(ctx context.Context, name string) ([]*Agent, error)
	UpdateLastConnected(ctx context.Context, id uuid.UUID) error
	UpdatePublicKey(ctx context.Context, id uuid.UUID, publicKey string) error
}

