package repository

import (
	"context"

	"github.com/NirajDonga/dbpods/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByOAuthID(ctx context.Context, oauthID string) (*models.User, error) {
	query := `SELECT id, email, oauth_id, created_at FROM users WHERE oauth_id = $1`

	var user models.User
	err := r.pool.QueryRow(ctx, query, oauthID).Scan(
		&user.ID, &user.Email, &user.OAuthID, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, email, oauthID string) (*models.User, error) {
	query := `
		INSERT INTO users (email, oauth_id) 
		VALUES ($1, $2) 
		RETURNING id, email, oauth_id, created_at`

	var user models.User
	err := r.pool.QueryRow(ctx, query, email, oauthID).Scan(
		&user.ID, &user.Email, &user.OAuthID, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
