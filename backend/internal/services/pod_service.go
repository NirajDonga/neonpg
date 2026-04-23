package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NirajDonga/dbpods/internal/kubernetes"
	"github.com/NirajDonga/dbpods/internal/models"
	"github.com/NirajDonga/dbpods/internal/repository"
)

type PodService struct {
	podRepo *repository.PodRepository
	k8sClient *kubernetes.Client
}

func NewPodService(podRepo *repository.PodRepository, k8sClient *kubernetes.Client) *PodService {
	return &PodService{podRepo: podRepo, k8sClient: k8sClient}
}

func generateSecurePassword() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *PodService) ProvisionDatabase(ctx context.Context, userID int) (*models.Pod, string, string, error) {
	// 1. Check if user already has a database (Quota limit of 1 for free tier)
	existingPods, err := s.podRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to check existing pods: %w", err)
	}

	if len(existingPods) > 0 {
		return nil, "", "", fmt.Errorf("free tier limit reached: you have already claimed a free database")
	}

	tenantID := fmt.Sprintf("tenant-db-%d-%d", userID, time.Now().Unix())

	dbPassword, err := generateSecurePassword()
	if err != nil {
		return nil, "", "", err
	}

	pod, err := s.podRepo.Create(ctx, userID, tenantID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to save pod to database: %w", err)
	}

	// 2. Provision the actual Kubernetes resources (Service + StatefulSet)
	err = s.k8sClient.CreatePostgresDeployment(ctx, tenantID, dbPassword)
	if err != nil {
		s.podRepo.UpdateStatus(ctx, pod.ID, "error")
		return nil, "", "", fmt.Errorf("k8s provisioning failed: %w", err)
	}

	// 3. Make the connection string (in-cluster DNS)
	connectionString := fmt.Sprintf("postgres://postgres:%s@%s-svc:5432/postgres", dbPassword, tenantID)

	return pod, connectionString, dbPassword, nil
}

func (s *PodService) GetUserPods(ctx context.Context, userID int) ([]models.Pod, error) {
	return s.podRepo.GetByUserID(ctx, userID)
}
