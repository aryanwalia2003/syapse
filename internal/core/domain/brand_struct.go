// internal/core/domain/brand_struct.go
package domain

import "time"

type Brand struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ContactEmail string    `json:"contact_email"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
