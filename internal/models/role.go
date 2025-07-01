package models

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Permissions []string  `json:"permissions" db:"permissions"`
	Description string    `json:"description" db:"description"`
}
