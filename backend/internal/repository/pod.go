package repository

import (
	"context"

	"github.com/NirajDonga/dbpods/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PodRepository struct {
	pool *pgxpool.Pool
}

func NewPodRepository(pool *pgxpool.Pool) *PodRepository {
	return &PodRepository{pool: pool}
}

func (r *PodRepository) Create(ctx context.Context, userID int, tenantID string) (*models.Pod, error) {
	query := `
		INSERT INTO pods (user_id, tenant_id, status, expires_at)
		VALUES ($1, $2, 'running', NOW() + INTERVAL '1 hour')
		RETURNING id, user_id, tenant_id, status, created_at, expires_at`

	var pod models.Pod
	err := r.pool.QueryRow(ctx, query, userID, tenantID).Scan(
		&pod.ID, &pod.UserID, &pod.TenantID, &pod.Status, &pod.CreatedAt, &pod.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func (r *PodRepository) GetByUserID(ctx context.Context, userID int) ([]models.Pod, error) {
	query := `
		SELECT id, user_id, tenant_id, status, created_at, expires_at 
		FROM pods 
		WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pods []models.Pod
	for rows.Next() {
		var pod models.Pod
		if err := rows.Scan(&pod.ID, &pod.UserID, &pod.TenantID, &pod.Status, &pod.CreatedAt, &pod.ExpiresAt); err != nil {
			return nil, err
		}
		pods = append(pods, pod)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pods, nil
}

func (r *PodRepository) UpdateStatus(ctx context.Context, podID int, status string) error {
	query := `UPDATE pods SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, podID)
	return err
}

func (r *PodRepository) GetExpiredPods(ctx context.Context) ([]models.Pod, error) {
	query := `
		SELECT id, user_id, tenant_id, status, created_at, expires_at 
		FROM pods 
		WHERE status = 'running' AND expires_at < NOW()`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pods []models.Pod
	for rows.Next() {
		var pod models.Pod
		if err := rows.Scan(&pod.ID, &pod.UserID, &pod.TenantID, &pod.Status, &pod.CreatedAt, &pod.ExpiresAt); err != nil {
			return nil, err
		}
		pods = append(pods, pod)
	}
	return pods, rows.Err()
}
