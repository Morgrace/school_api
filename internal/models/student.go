package models

type Student struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty" validate:"required"`
	LastName  string `json:"last_name,omitempty" validate:"required"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Class     string `json:"class,omitempty" validate:"required"`
}

type StudentFilter struct {
	FirstName string
	LastName  string
	Email     string
	Class     string

	SortBy    string // e.g. "email"
	SortOrder string // e.g. "ASC" or "DESC"
}
