package models

type Teacher struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty" validate:"required"`
	LastName  string `json:"last_name,omitempty" validate:"required"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Class     string `json:"class,omitempty" validate:"required"`
	Subject   string `json:"subject,omitempty" validate:"required"`
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
