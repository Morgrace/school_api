package router

import (
	"net/http"
	// Make sure this matches your folder structure.
	// If handlers are in "simpleapi/internal/handlers", use that.
	"simpleapi/internal/api/handlers"
)

// Router now accepts the fully initialized TeacherHandler
func Router(h *handlers.TeacherHandler) *http.ServeMux {

	mux := http.NewServeMux()

	// If RootHandler is still a simple function (not in the struct),
	// you can call it directly like this:
	// mux.HandleFunc("/", handlers.RootHandler)

	// --- COLLECTION ROUTES ---
	mux.HandleFunc("GET /teachers", h.GetTeachers)
	mux.HandleFunc("POST /teachers", h.CreateTeachers)       // renamed from AddTeacherHandler
	mux.HandleFunc("PATCH /teachers", h.BulkPatchTeachers)   // renamed from PatchTeachersHandler
	mux.HandleFunc("DELETE /teachers", h.BulkDeleteTeachers) // renamed from DeleteTeachersHandler

	// --- SINGLE ITEM ROUTES ---
	mux.HandleFunc("GET /teachers/{id}", h.GetTeacherByID)    // renamed from GetTeacherHandler
	mux.HandleFunc("PUT /teachers/{id}", h.UpdateTeacherFull) // renamed from UpdateTeacherHandler
	mux.HandleFunc("PATCH /teachers/{id}", h.PatchTeacher)    // renamed from PatchTeacherHandler
	mux.HandleFunc("DELETE /teachers/{id}", h.DeleteTeacher)  // renamed from DeleteOneTeacherHandler

	return mux
}
