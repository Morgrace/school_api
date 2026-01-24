package models

type Teacher struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Class     string `json:"class,omitempty"`
	Subject   string `json:"subject,omitempty"`
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
