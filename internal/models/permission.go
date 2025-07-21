package models

import "github.com/google/uuid"

type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description" db:"description"`
}
