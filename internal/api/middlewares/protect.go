package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"simpleapi/internal/repository" // Import your repo
	"simpleapi/pkg/utils"
)

// DENY BY DEFAULT FIXME ------------------ REMEMBER TO IMPLEMENT THIS PATTERN
// We define a key to store the FULL User object, not just the ID
type contextKey string

const UserKey contextKey = "currentUser"

// AuthMiddleware holds the dependencies (The Database Repo)
type AuthMiddleware struct {
	Repo *repository.TeacherRepository
}

// NewAuthMiddleware is the constructor
func NewAuthMiddleware(repo *repository.TeacherRepository) *AuthMiddleware {
	return &AuthMiddleware{Repo: repo}
}

// Protect is the actual middleware function (mirrors your TS 'protect')
func (m *AuthMiddleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// 1. EXTRACT TOKEN (Hybrid: Cookie or Header)
		// Check Cookie first (Web Client)
		if cookie, err := r.Cookie("session_token"); err == nil {
			tokenString = cookie.Value
		}

		// If no cookie, check Header (Mobile/API Client)
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// If still empty -> 401
		if tokenString == "" {
			utils.WriteError(w, http.StatusUnauthorized, "You are not logged in!")
			return
		}

		// 2. VALIDATE TOKEN (Check Signature)
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// 3. FETCH USER FROM DB (The "Robust" Step)
		// We use the ID from the token claims to find the real user
		// Note: claims.Subject is usually a string, convert if your ID is int
		userID, _ := strconv.Atoi(claims.UserID)

		currentUser, err := m.Repo.GetByID(r.Context(), userID)
		if err != nil {
			// If error is "No Rows Found", it means User was DELETED
			utils.WriteError(w, http.StatusUnauthorized, "The user belonging to this token no longer exists.")
			return
		}

		// 4. CHECK IF PASSWORD CHANGED (Security Critical)
		// Compare "Token Issue Date" (iat) vs "Password Changed Date"
		// Note: You need to implement ChangedPasswordAfter in your model or helper
		// valid: check if IssuedAt is not nil to avoid panic
		if claims.IssuedAt != nil {
			// Extract the .Time (Go Time object) and convert to .Unix() (int64)
			if currentUser.ChangedPasswordAfter(claims.IssuedAt.Time.Unix()) {
				utils.WriteError(w, http.StatusUnauthorized, "User recently changed password! Please log in again.")
				return
			}
		}
		// 5. SUCCESS: Attach the FULL User to Context
		// Now handlers don't need to query the DB anymore!
		ctx := context.WithValue(r.Context(), UserKey, currentUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
