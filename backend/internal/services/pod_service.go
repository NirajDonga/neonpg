package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NirajDonga/dbpods/internal/models"
	"github.com/NirajDonga/dbpods/internal/repository"
)

type PodService struct {
	podRepo *repository.PodRepository
}

func NewPodService(podRepo *repository.PodRepository) *PodService {
	return &PodService{podRepo: podRepo}
}

// generateSecurePassword creates a randomized 16-character string
func generateSecurePassword() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ProvisionDatabase handles all the business logic for creating a new database pod
func (s *PodService) ProvisionDatabase(ctx context.Context, userID int) (*models.Pod, string, string, error) {
	tenantID := fmt.Sprintf("tenant-db-%d-%d", userID, time.Now().Unix())
	
	dbPassword, err := generateSecurePassword()
	if err != nil {
		return nil, "", "", err
	}

	// 1. Record the new pod in the database
	pod, err := s.podRepo.Create(ctx, userID, tenantID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to save pod to database: %w", err)
	}

	// 2. TODO: Infrastructure Provisioning
	// Call Kubernetes/Docker SDK here
	// err = k8sClient.CreatePostgresDeployment(tenantID, dbPassword)
	// if err != nil { 
	//     s.podRepo.UpdateStatus(ctx, pod.ID, "error")
	//     return nil, "", "", err 
	// }

	// 3. Make the connection string
	connectionString := fmt.Sprintf("postgres://postgres:%s@%s.dbpods.internal:5432/postgres", dbPassword, tenantID)

	return pod, connectionString, dbPassword, nil
}

// GetUserPods is just a pass-through to the repository, but belongs in the service layer
func (s *PodService) GetUserPods(ctx context.Context, userID int) ([]models.Pod, error) {
	return s.podRepo.GetByUserID(ctx, userID)
}
