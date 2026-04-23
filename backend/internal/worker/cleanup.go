package worker

import (
	"context"
	"log"
	"time"

	"github.com/NirajDonga/dbpods/internal/kubernetes"
	"github.com/NirajDonga/dbpods/internal/repository"
)

type CleanupWorker struct {
	podRepo   *repository.PodRepository
	k8sClient *kubernetes.Client
	interval  time.Duration
}

func NewCleanupWorker(podRepo *repository.PodRepository, k8sClient *kubernetes.Client) *CleanupWorker {
	return &CleanupWorker{
		podRepo:   podRepo,
		k8sClient: k8sClient,
		interval:  60 * time.Second,
	}
}

func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("[CleanupWorker] Started. Checking for expired pods every 60 seconds.")

	for {
		select {
		case <-ticker.C:
			w.run(ctx)
		case <-ctx.Done():
			log.Println("[CleanupWorker] Shutting down.")
			return
		}
	}
}

func (w *CleanupWorker) run(ctx context.Context) {
	expiredPods, err := w.podRepo.GetExpiredPods(ctx)
	if err != nil {
		log.Printf("[CleanupWorker] Failed to fetch expired pods: %v", err)
		return
	}

	if len(expiredPods) == 0 {
		return
	}

	log.Printf("[CleanupWorker] Found %d expired pod(s) to clean up.", len(expiredPods))

	for _, pod := range expiredPods {
		if err := w.k8sClient.DeletePostgresDeployment(ctx, pod.TenantID); err != nil {
			log.Printf("[CleanupWorker] Failed to delete k8s resources for tenant %s: %v", pod.TenantID, err)
			continue
		}

		if err := w.podRepo.UpdateStatus(ctx, pod.ID, "expired"); err != nil {
			log.Printf("[CleanupWorker] Failed to update status for pod %d: %v", pod.ID, err)
			continue
		}

		log.Printf("[CleanupWorker] Successfully cleaned up pod for tenant: %s", pod.TenantID)
	}
}
