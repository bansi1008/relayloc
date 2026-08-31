package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, agent *Agent) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO agents (
			id,
			user_id,
			name,
			created_at,
			updated_at,
			last_connected_at,
			credential_hash
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		agent.ID,
		agent.UserID,
		agent.Name,
		agent.CreatedAt,
		agent.UpdatedAt,
		agent.LastConnectedAt,
		agent.Token,
	)

	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*Agent, error) {

	rows, err := r.db.Query(
		ctx,
		`
        SELECT
            id,
            user_id,
            name,
            created_at,
            updated_at,
            last_connected_at
        FROM agents
        WHERE user_id = $1
        `,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	agents := []*Agent{}

	for rows.Next() {
		a := &Agent{}
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Name,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.LastConnectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}

		agents = append(agents, a)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return agents, nil
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Agent, error) {
	a := &Agent{}
	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			name,
			COALESCE(public_key, ''),
			created_at,
			updated_at,
			last_connected_at
		FROM agents
		WHERE id = $1
		`,
		id,
	).Scan(
		&a.ID,
		&a.UserID,
		&a.Name,
		&a.PublicKey,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.LastConnectedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("get agent by id: %w", err)
	}

	return a, nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	result, err := r.db.Exec(
		ctx,
		`DELETE FROM agents WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("agent not found")
	}

	return nil
}

func (r *PostgresRepository) GetByName(
	ctx context.Context,
	name string,
	userID uuid.UUID,
) (*Agent, error) {
	a := &Agent{}
	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			name,
			credential_hash,
			COALESCE(public_key, ''),
			created_at,
			updated_at,
			last_connected_at
		FROM agents
		WHERE name = $1 AND user_id=$2
		`,
		name,
		userID,
	).Scan(
		&a.ID,
		&a.UserID,
		&a.Name,
		&a.Token,
		&a.PublicKey,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.LastConnectedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("get agent by id: %w", err)
	}

	return a, nil
}

func (r *PostgresRepository) GetByNameOnly(
	ctx context.Context,
	name string,
) ([]*Agent, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			user_id,
			name,
			credential_hash,
			created_at,
			updated_at,
			last_connected_at
		FROM agents
		WHERE name = $1
		`,
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("query agents by name: %w", err)
	}
	defer rows.Close()

	agents := []*Agent{}
	for rows.Next() {
		a := &Agent{}
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Name,
			&a.Token,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.LastConnectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, a)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return agents, nil
}

func (r *PostgresRepository) UpdateLastConnected(
	ctx context.Context,
	id uuid.UUID,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE agents SET last_connected_at = NOW(), updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("update last connected: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdatePublicKey(
	ctx context.Context,
	id uuid.UUID,
	publicKey string,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE agents SET public_key = $1, updated_at = NOW() WHERE id = $2`,
		publicKey,
		id,
	)
	if err != nil {
		return fmt.Errorf("update public key: %w", err)
	}
	return nil
}


