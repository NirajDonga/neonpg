package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/NirajDonga/dbpods/internal/core"
)

type APIHandler struct {
	provisioner core.DBProvisioner
}

func NewAPIHandler(p core.DBProvisioner) *APIHandler {
	return &APIHandler{provisioner: p}
}

type CreateDBRequest struct {
	TenantID string `json:"tenantId"`
}

func (h *APIHandler) HandleCreateDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dbPassword := fmt.Sprintf("secure_%d", rng.Intn(999999))

	err := h.provisioner.CreateTenantDatabase(r.Context(), req.TenantID, dbPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to provision DB: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Database provisioning successfully started",
		"tenant":  req.TenantID,
	})
}
