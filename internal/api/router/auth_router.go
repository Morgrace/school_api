package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
)

func authenticationRoutes(mux *http.ServeMux, h *handlers.TeacherHandler) {
	mux.HandleFunc("POST /login", h.LoginTeacher)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("POST /register", h.RegisterTeacher)
	// mux.HandleFunc("PATCH /update-password")
	// mux.HandleFunc("POST /forgot-password")
	// mux.HandleFunc("POST /reset-password/{reset-token}")
}
