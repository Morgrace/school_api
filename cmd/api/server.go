package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	mw "simpleapi/internal/api/middlewares"
	"strings"
	"time"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET method blah blah balh"))
	case http.MethodPost:
		w.Write([]byte("Hello POST method"))
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH method"))
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method"))
	case http.MethodPut:
		w.Write([]byte("Hello PUT method"))
	}
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	userID := strings.TrimSuffix(path, "/")
	fmt.Println(userID)
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET method"))
	case http.MethodPost:
		w.Write([]byte("Hello POST method"))
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH method"))
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method"))
	case http.MethodPut:
		w.Write([]byte("Hello PUT method"))
	}
}
func main() {
	port := ":3000"
	cert := "cert.pem"
	key := "key.pem"
	mux := http.NewServeMux()
	// ROUTE HANDLERS
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/teachers", teachersHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := mw.NewRateLimiter(5, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		Whitelist:                   []string{"sortBy","sortOrder","name","age","class"},
	}
	// Create custom server
	server := &http.Server{
		Addr:      port,
		Handler:   mw.Hpp(hppOptions)(rl.Middleware(mw.Compression(mw.ResponseTimeMiddleware(mw.SecurityHeaders(mw.Cors(mux)))))),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
