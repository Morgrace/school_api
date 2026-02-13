package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
	mw "simpleapi/internal/api/middlewares"
)

func registerTeachersRoutes(mux *http.ServeMux, h *handlers.TeacherHandler, am *mw.AuthMiddleware) {
	protect := func(next http.HandlerFunc) http.Handler {
		return am.Protect(next)
	}
	mux.Handle("GET /teachers", protect(h.GetTeachers))
	mux.Handle("POST /teachers", protect(h.CreateTeachers))
	mux.HandleFunc("PATCH /teachers", h.BulkPatchTeachers)
	mux.HandleFunc("DELETE /teachers", h.BulkDeleteTeachers)
	mux.HandleFunc("GET /teachers/{id}", h.GetTeacherByID)
	mux.HandleFunc("PUT /teachers/{id}", h.UpdateTeacherFull)
	mux.HandleFunc("PATCH /teachers/{id}", h.PatchTeacher)
	mux.HandleFunc("DELETE /teachers/{id}", h.DeleteTeacher)

	mux.HandleFunc("GET /teachers/{id}/students", h.GetStudentsByTeacherId)
	mux.HandleFunc("GET /teachers/{id}/studentCount", h.GetStudentsByTeacherId)
}
