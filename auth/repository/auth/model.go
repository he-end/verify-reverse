package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	ID            uuid.UUID  `bun:"id,pk,type:uuid"`
	Number        *string    `bun:"number,unique,nullzero"`
	Email         *string    `bun:"email,unique,nullzero"`
	Name          string     `bun:"name,notnull"`
	PasswordHash  string     `bun:"password_hash,nullzero"`
	Status        string     `bun:"status,notnull"`
	VerifiedAt    *time.Time `bun:"verified_at"`
	CreatedAt     time.Time  `bun:"created_at,notnull"`
	UpdatedAt     time.Time  `bun:"updated_at,notnull"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Number: u.Number,
		Status: u.Status,
	}
}

type UserResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  *string   `json:"email,omitempty"`
	Number *string   `json:"number,omitempty"`
	Status string    `json:"status"`
}
