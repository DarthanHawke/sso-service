package models

type Role struct {
	ID          string   `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Permissions []string `json:"permissions" db:"permissions"`
	Description string   `json:"description" db:"description"`
}
