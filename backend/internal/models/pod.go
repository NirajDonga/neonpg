package models

import "time"

type Pod struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	TenantID  string    `json:"tenant_id"` 
	Status    string    `json:"status"`   
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
