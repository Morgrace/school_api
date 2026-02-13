package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
)

func registerStudentRoutes(mux *http.ServeMux, h *handlers.StudentHandler) {
	mux.HandleFunc("GET /students", h.GetStudents)
	mux.HandleFunc("POST /students", h.CreateStudents)
}
