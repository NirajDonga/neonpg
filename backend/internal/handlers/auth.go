package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/NirajDonga/dbpods/internal/config"
	"github.com/NirajDonga/dbpods/internal/repository"
	"github.com/NirajDonga/dbpods/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	cfg         *config.AppConfig
	userRepo    *repository.UserRepository
	oauthConfig *oauth2.Config
}

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func NewAuthHandler(cfg *config.AppConfig, userRepo *repository.UserRepository) *AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleCallbackURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
	return &AuthHandler{cfg: cfg, userRepo: userRepo, oauthConfig: oauthConfig}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info from Google"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info response"})
		return
	}

	var userInfo googleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
		return
	}

	ctx := c.Request.Context()

	user, err := h.userRepo.GetByOAuthID(ctx, userInfo.Sub)
	if err != nil {
		user, err = h.userRepo.Create(ctx, userInfo.Email, userInfo.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	jwtToken, err := utils.GenerateToken(user.ID, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   jwtToken,
		"user":    user,
	})
}
