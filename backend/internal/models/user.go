package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	OAuthID   string    `json:"oauth_id"`
	CreatedAt time.Time `json:"created_at"`
}
