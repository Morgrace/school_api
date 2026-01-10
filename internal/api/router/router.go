package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
)

func Router() *http.ServeMux {

	mux := http.NewServeMux()
	// ROUTE HANDLERS
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers/", handlers.TeachersHandler)
	
	return mux
}
