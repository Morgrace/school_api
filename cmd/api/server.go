package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"simpleapi/internal/api/handlers"
	mw "simpleapi/internal/api/middlewares"
	"simpleapi/internal/api/router"
	"simpleapi/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Config
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env")
	}

	// 2. Initialize Database (The Pro Way: returns the instance, no global var)
	db, err := repository.NewDB()
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer db.Close() // Main owns the cleanup

	// 3. WIRING: Dependency Injection Chain
	// Level 1: Create the Repository (injects DB)
	teacherRepo := repository.NewTeacherRepository(db)

	// Level 2: Create the Handler (injects Repo)
	teacherHandler := handlers.NewTeacherHandler(teacherRepo)

	// Level 3: Create the Router (injects Handler)
	// Note: We need to update your router.Router() function to accept this argument!
	mux := router.Router(teacherHandler)

	port := os.Getenv("SERVER_PORT")

	cert := "cert.pem"
	key := "key.pem"

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// rl := mw.NewRateLimiter(5, time.Minute)
	// hppOptions := mw.HPPOptions{
	// 	CheckQuery:                  true,
	// 	CheckBody:                   true,
	// 	CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
	// 	Whitelist:                   []string{"sortBy", "sortOrder", "name", "age", "class"},
	// }
	// secureMux := mw.Cors(rl.Middleware(mw.ResponseTimeMiddleware(mw.SecurityHeaders(mw.Compression(mw.Hpp(hppOptions)(mux))))))
	// secureMux:= applyMiddlewares(mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware, rl.Middleware, mw.Cors)
	secureMux := mw.SecurityHeaders(mux)
	// Create custom server
	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
