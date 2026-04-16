package handlers

import (
	"net/http"

	"github.com/NirajDonga/dbpods/internal/services"
	"github.com/gin-gonic/gin"
)

type PodHandler struct {
	podService *services.PodService
}

func NewPodHandler(podService *services.PodService) *PodHandler {
	return &PodHandler{podService: podService}
}

func (h *PodHandler) CreatePod(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}
	userID := userIDRaw.(int)

	// Delegate all complex logic to the Service Layer
	pod, connString, password, err := h.podService.ProvisionDatabase(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "pod created successfully",
		"pod":               pod,
		"connection_string": connString,
		"password":          password,
	})
}

func (h *PodHandler) GetUserPods(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}
	userID := userIDRaw.(int)

	pods, err := h.podService.GetUserPods(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pods"})
		return
	}

	if pods == nil {
		c.JSON(http.StatusOK, gin.H{"pods": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pods": pods})
}
