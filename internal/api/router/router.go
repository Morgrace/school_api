package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
	"simpleapi/internal/api/middlewares"
)

func Router(th *handlers.TeacherHandler, sh *handlers.StudentHandler, am *middlewares.AuthMiddleware) *http.ServeMux {
	// 1. Create the Main Traffic Controller
	mainMux := http.NewServeMux()

	// 2. Create the V1 Sub-Router (The Shared Canvas)
	v1 := http.NewServeMux()

	// 3. Hand the V1 canvas to your sub-routers to paint their routes
	authenticationRoutes(v1, th)
	registerTeachersRoutes(v1, th,am)
	registerStudentRoutes(v1, sh)

	// 4. Mount the filled-up V1 router onto the main router
	// Any request starting with "/api/v1/" gets stripped and sent to 'v1'
	mainMux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	return mainMux
}
