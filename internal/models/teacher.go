package models

import (
	"time"
)

type Teacher struct {
	// -- CORE IDENTITY FIELDS --
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty" validate:"required"`
	LastName  string `json:"last_name,omitempty" validate:"required"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Role      string `json:"role"`
	// --- SCHOOL DATA FIELDS ---
	Class   string `json:"class,omitempty" validate:"required"`
	Subject string `json:"subject,omitempty" validate:"required"`
	// --- SECURITY & ACCOUNT FIELDS ---
	Password     string `json:"password,omitempty" validate:"required"`
	PasswordHash string `json:"-"`

	PasswordChangedAt    *time.Time `json:"-"`
	PasswordResetToken   *string    `json:"-"`
	PasswordResetExpires *time.Time `json:"-"`

	// --- META FIELDS ---
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	IsActive  bool       `json:"is_active"`
}

// TeacherFilter allows the Handler to tell the Repo what to search for
// without passing the raw *http.Request
type TeacherFilter struct {
	FirstName string
	LastName  string
	Email     string
	Class     string
	Subject   string

	SortBy    string // e.g. "email"
	SortOrder string // e.g. "ASC" or "DESC"
}

// Internal/models/teacher.go

func (t *Teacher) ChangedPasswordAfter(jwtTimestamp int64) bool {
	// 1. Safety Check: If user never changed password, return false
	// (Assuming PasswordChangedAt is a *time.Time or time.Time.
	// If it's a pointer, check for nil first)
	if t.PasswordChangedAt == nil {
		return false
	}

	if t.PasswordChangedAt.IsZero() {
		return false
	}

	// 2. Convert the Database Time to Unix Timestamp (int64)
	// We strictly compare seconds.
	passwordChangedTimestamp := t.PasswordChangedAt.Unix()

	// 3. The Comparison
	// If the password was changed AFTER the token was issued,
	// then changedTimestamp will be LARGER than jwtTimestamp.
	//
	// Example:
	// Token Issued: 10:00 AM (Timestamp: 1000)
	// Pass Changed: 10:05 AM (Timestamp: 1005)
	// 1005 > 1000 -> TRUE (Reject the token)

	return passwordChangedTimestamp > jwtTimestamp
}
