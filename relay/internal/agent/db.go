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
			connected
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		agent.ID,
		agent.UserID,
		agent.Name,
		agent.CreatedAt,
		agent.UpdatedAt,
		agent.LastConnectedAt,
		agent.Connected,
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
            last_connected_at,
            connected
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
			&a.Connected,
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
			created_at,
			updated_at,
			last_connected_at,
			connected
		FROM agents
		WHERE id = $1
		`,
		id,
	).Scan(
		&a.ID,
		&a.UserID,
		&a.Name,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.LastConnectedAt,
		&a.Connected,
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

