package router

import (
	"net/http"
	// Make sure this matches your folder structure.
	// If handlers are in "simpleapi/internal/handlers", use that.
	"simpleapi/internal/api/handlers"
)

func Router(th *handlers.TeacherHandler, sh *handlers.StudentHandler) *http.ServeMux {

	// 1. Create the Sub-Router for V1
	// This router thinks it is at the root ("/")
	v1 := http.NewServeMux()

	// --- TEACHER ROUTES ---
	v1.HandleFunc("GET /teachers", th.GetTeachers)
	v1.HandleFunc("POST /teachers", th.CreateTeachers)
	v1.HandleFunc("PATCH /teachers", th.BulkPatchTeachers)
	v1.HandleFunc("DELETE /teachers", th.BulkDeleteTeachers)
	v1.HandleFunc("GET /teachers/{id}", th.GetTeacherByID)
	v1.HandleFunc("PUT /teachers/{id}", th.UpdateTeacherFull)
	v1.HandleFunc("PATCH /teachers/{id}", th.PatchTeacher)
	v1.HandleFunc("DELETE /teachers/{id}", th.DeleteTeacher)

	// --- SUB ROUTES ---
	v1.HandleFunc("GET /teachers/{id}/students", th.GetStudentsByTeacherId)
	v1.HandleFunc("GET /teachers/{id}/studentCount", th.GetStudentsByTeacherId)

	// ---- STUDENT ROUTES ---
	v1.HandleFunc("GET /students", sh.GetStudents)
	v1.HandleFunc("POST /students", sh.CreateStudents)

	mainMux := http.NewServeMux()

	mainMux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	return mainMux
}
